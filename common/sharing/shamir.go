//
// Copyright Coinbase, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0
//

// Package sharing is an implementation of shamir secret sharing and implements the following papers.
//
// - https://dl.acm.org/doi/pdf/10.1145/359168.359176
// - https://www.cs.umd.edu/~gasarch/TOPICS/secretsharing/feldmanVSS.pdf
// - https://link.springer.com/content/pdf/10.1007%2F3-540-46766-1_9.pdf
package sharing

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/sharing"
	log "github.com/sirupsen/logrus"
)

type ShamirShare struct {
	Id    uint32 `json:"identifier"`
	Value []byte `json:"value"`
}

func (ss ShamirShare) Validate(curve *curves.Curve) error {
	if ss.Id == 0 {
		return fmt.Errorf("invalid identifier")
	}
	sc, err := curve.Scalar.SetBytes(ss.Value)
	if err != nil {
		return err
	}
	if sc.IsZero() {
		return fmt.Errorf("invalid share")
	}
	return nil
}

func (ss ShamirShare) Bytes() []byte {
	var id [4]byte
	binary.BigEndian.PutUint32(id[:], ss.Id)
	return append(id[:], ss.Value...)
}

type Shamir struct {
	threshold, limit uint32
	curve            *curves.Curve
}

func NewShamir(threshold, limit uint32, curve *curves.Curve) (*Shamir, error) {
	if limit < threshold {
		return nil, fmt.Errorf("limit cannot be less than threshold")
	}
	if threshold < 2 {
		return nil, fmt.Errorf("threshold cannot be less than 2")
	}
	if limit > 255 {
		return nil, fmt.Errorf("cannot exceed 255 shares")
	}
	if curve == nil {
		return nil, fmt.Errorf("invalid curve")
	}
	return &Shamir{threshold, limit, curve}, nil
}

func (s Shamir) Split(secret curves.Scalar, reader io.Reader) ([]*ShamirShare, error) {
	if secret.IsZero() {
		return nil, fmt.Errorf("invalid secret")
	}
	shares, _ := s.getPolyAndShares(secret, reader)
	return shares, nil
}

func (s Shamir) getPolyAndShares(secret curves.Scalar, reader io.Reader) ([]*ShamirShare, *sharing.Polynomial) {
	poly := new(sharing.Polynomial).Init(secret, s.threshold, reader)
	shares := make([]*ShamirShare, s.limit)
	for i := range shares {
		x := s.curve.Scalar.New(i + 1)
		shares[i] = &ShamirShare{
			Id:    uint32(i + 1),
			Value: poly.Evaluate(x).Bytes(),
		}
	}
	return shares, poly
}

// TODO this is duplicate code, but should be here in common folder
func verifierFromCommits(k int, c []byte, curve *curves.Curve) (*sharing.FeldmanVerifier, error) {

	commitment, err := DecompressCommitments(k, c, curve)
	if err != nil {
		return nil, err
	}
	verifier := new(sharing.FeldmanVerifier)
	verifier.Commitments = commitment
	return verifier, nil
}

// TODO this is (almost) duplicate code, but should be here in common folder
// Predicate verifies if the share fits the polynomial commitments
func Predicate(secret_point curves.Point, cipher []byte, commits []byte, k int, curve *curves.Curve) (*sharing.ShamirShare, *sharing.FeldmanVerifier, bool) {
	keyBytes := secret_point.ToAffineCompressed()

	// Hash the serialized point to derive a symmetric key.
	hashedKey := sha256.Sum256(keyBytes)

	shareBytes, err := Decrypt(hashedKey, cipher)
	if err != nil {
		log.Errorf("Error while decrypting share: err=%s", err)
		return nil, nil, false
	}
	share := sharing.ShamirShare{Id: binary.BigEndian.Uint32(shareBytes[:4]), Value: shareBytes[4:]}
	log.Debugf("share: id=%d, val=%v", share.Id, share.Value)
	verifier, err := verifierFromCommits(k, commits, curve)
	if err != nil {
		log.Errorf("Error while getting verifier from commits=%s", err)
		return nil, nil, false
	}

	if err = verifier.Verify(&share); err != nil {
		log.Errorf("Error while verifying share=%s", err)
		return nil, nil, false
	}
	return &share, verifier, true
}

func (s Shamir) LagrangeCoeffs(identities []uint32) (map[uint32]curves.Scalar, error) {
	xs := make(map[uint32]curves.Scalar, len(identities))
	for _, xi := range identities {
		xs[xi] = s.curve.Scalar.New(int(xi))
	}

	result := make(map[uint32]curves.Scalar, len(identities))
	for i, xi := range xs {
		num := s.curve.Scalar.One()
		den := s.curve.Scalar.One()
		for j, xj := range xs {
			if i == j {
				continue
			}

			num = num.Mul(xj)
			den = den.Mul(xj.Sub(xi))
		}
		if den.IsZero() {
			return nil, fmt.Errorf("divide by zero")
		}
		result[i] = num.Div(den)
	}
	return result, nil
}

func (s Shamir) Combine(shares ...*ShamirShare) (curves.Scalar, error) {
	if len(shares) < int(s.threshold) {
		return nil, fmt.Errorf("invalid number of shares")
	}
	dups := make(map[uint32]bool, len(shares))
	xs := make([]curves.Scalar, len(shares))
	ys := make([]curves.Scalar, len(shares))

	for i, share := range shares {
		err := share.Validate(s.curve)
		if err != nil {
			return nil, err
		}
		if share.Id > s.limit {
			return nil, fmt.Errorf("invalid share identifier")
		}
		if _, in := dups[share.Id]; in {
			return nil, fmt.Errorf("duplicate share")
		}
		dups[share.Id] = true
		ys[i], _ = s.curve.Scalar.SetBytes(share.Value)
		xs[i] = s.curve.Scalar.New(int(share.Id))
	}
	return s.interpolate(xs, ys)
}

func (s Shamir) CombinePoints(shares ...*ShamirShare) (curves.Point, error) {
	if len(shares) < int(s.threshold) {
		return nil, fmt.Errorf("invalid number of shares")
	}
	dups := make(map[uint32]bool, len(shares))
	xs := make([]curves.Scalar, len(shares))
	ys := make([]curves.Point, len(shares))

	for i, share := range shares {
		err := share.Validate(s.curve)
		if err != nil {
			return nil, err
		}
		if share.Id > s.limit {
			return nil, fmt.Errorf("invalid share identifier")
		}
		if _, in := dups[share.Id]; in {
			return nil, fmt.Errorf("duplicate share")
		}
		dups[share.Id] = true
		sc, _ := s.curve.Scalar.SetBytes(share.Value)
		ys[i] = s.curve.ScalarBaseMult(sc)
		xs[i] = s.curve.Scalar.New(int(share.Id))
	}
	return s.interpolatePoint(xs, ys)
}

func (s Shamir) interpolate(xs, ys []curves.Scalar) (curves.Scalar, error) {
	result := s.curve.Scalar.Zero()
	for i, xi := range xs {
		num := s.curve.Scalar.One()
		den := s.curve.Scalar.One()
		for j, xj := range xs {
			if i == j {
				continue
			}
			num = num.Mul(xj)
			den = den.Mul(xj.Sub(xi))
		}
		if den.IsZero() {
			return nil, fmt.Errorf("divide by zero")
		}
		result = result.Add(ys[i].Mul(num.Div(den)))
	}
	return result, nil
}

// interpolateShares interpolates the shares at a given x value and returns the result as curves.Scalar.
func (s *Shamir) ObtainEvalForX(shares []*ShamirShare, xValue uint32) (curves.Scalar, error) {
	x := s.curve.Scalar.New(int(xValue))
	result := s.curve.Scalar.Zero()

	for i, shareI := range shares {
		xi := s.curve.Scalar.New(int(shareI.Id))
		num := s.curve.Scalar.One()
		den := s.curve.Scalar.One()

		for j, shareJ := range shares {
			if i == j {
				continue
			}

			xj := s.curve.Scalar.New(int(shareJ.Id))
			num = num.Mul(x.Sub(xj))  // (x - xj)
			den = den.Mul(xi.Sub(xj)) // (xi - xj)
		}

		if den.IsZero() {
			return nil, fmt.Errorf("divide by zero in interpolation")
		}

		// Convert share value to curve's scalar
		shareValue, err := s.curve.Scalar.SetBytes(shareI.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to set share value to curve's scalar: %v", err)
		}

		// Calculate the term for this share and add it to the result
		term := shareValue.Mul(num).Div(den)
		result = result.Add(term)
	}

	return result, nil
}

func (s Shamir) interpolatePoint(xs []curves.Scalar, ys []curves.Point) (curves.Point, error) {
	result := s.curve.NewIdentityPoint()
	for i, xi := range xs {
		num := s.curve.Scalar.One()
		den := s.curve.Scalar.One()
		for j, xj := range xs {
			if i == j {
				continue
			}
			num = num.Mul(xj)
			den = den.Mul(xj.Sub(xi))
		}
		if den.IsZero() {
			return nil, fmt.Errorf("divide by zero")
		}
		result = result.Add(ys[i].Mul(num.Div(den)))
	}
	return result, nil
}

// GenerateCommitmentAndShares generates a commitment and shares for a given secret
// using the provided parameters.
func GenerateCommitmentAndShares(secret curves.Scalar, k, n uint32, curve *curves.Curve) (*sharing.FeldmanVerifier, []*sharing.ShamirShare, error) {
	f, err := sharing.NewFeldman(k, n, curve)
	if err != nil {
		return nil, nil, fmt.Errorf("gen_commitment_and_shares: %w", err)
	}

	feldcommit, shares, err := f.Split(secret, rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("gen_commitment_and_shares: %w", err)
	}
	return feldcommit, shares, nil
}
