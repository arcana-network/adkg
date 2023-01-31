package acss

type CommitmentScheme interface {
	Encrypt()
	CompressCommitments()
	DecompressCommitments()
	GenerateKeyPair()
	GenerateSecret()
	GenerateCommitmentAndShares()
	Split()
	SharedKey()
	Predicate()
	Encode()
	Decode()
}
