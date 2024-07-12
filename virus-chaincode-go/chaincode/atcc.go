package chaincode

import (
	"encoding/json"
	"fmt"
	"time"
	"log"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an signature
type VirusChaincode struct {
	contractapi.Contract
}

// signature describes basic details of what makes up a simple signature
// Insert struct field in alphabetic order => to achieve determinism accross languages
// golang keeps the order when marshal to json but doesn't order automatically
type VirusSignature struct {
  DocType     string `json:"docType"`
	IPFS_CID    string `json:"IPFS_CID"`
	SigName     string `json:"SigName"`
	Timestamp   int64  `json:"Timestamp"`
	Uploader    string `json:"Uploader"`
}

// Vote structure to hold vote details
type Vote struct {
  DocType string `json:"docType"`
	Org     string `json:"org"`
	Approve bool   `json:"approve"`
	SignatureCID string   `json:"signature_cid"`
}

func (t *VirusChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {

	virusSignatures := []VirusSignature{
		{IPFS_CID: "QmbR3sVeNPd1hp12sDh8Vz1BK7wretTenBe1sq5owNmDdB", SigName: "hypatia-md5-bloom", Timestamp: time.Now().Unix(), Uploader: "Divested-Mobile"},
		{IPFS_CID: "QmX75956Nbn5nVJzuup2zTwZiAMpoRYjV2WZe5PVziGwrk", SigName: "hypatia-sha1-bloom", Timestamp: time.Now().Unix(), Uploader: "Divested-Mobile"},
		{IPFS_CID: "QmeYqDrrqDUANAB1vWE4gjW85ZBmow97jAL9rFAM8eSwBQ", SigName: "hypatia-sha256-bloom", Timestamp: time.Now().Unix(), Uploader: "Divested-Mobile"},
		// Add more sample virus signatures as needed
	}

		for _, virusSignature := range virusSignatures {
		err := t.UploadSignature(ctx, virusSignature.IPFS_CID, virusSignature.SigName, virusSignature.Uploader)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *VirusChaincode) UploadSignature(ctx contractapi.TransactionContextInterface, ipfs_cid string, signame string, uploader string) error {
	exists, err := t.SignatureExists(ctx, ipfs_cid)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the signature %s already exists", ipfs_cid)
	}
	signature := VirusSignature{
    DocType: "signature",
		IPFS_CID:    ipfs_cid,
		SigName:   signame,
		Timestamp:   time.Now().Unix(),
		Uploader:    uploader,
	}

	sigJSON, err := json.Marshal(signature)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(ipfs_cid, sigJSON)
}

func (t *VirusChaincode) SignatureExists(ctx contractapi.TransactionContextInterface, ipfs_cid string) (bool, error) {
	// Check if the virus signature exists
	sigBytes, err := ctx.GetStub().GetState(ipfs_cid)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return sigBytes != nil, nil
}

func (t *VirusChaincode) QueryLatestSignatureBySigName(ctx contractapi.TransactionContextInterface, sigName string) ([]*VirusSignature, error) {
	// Construct the query string
	queryString := fmt.Sprintf(`{
		"selector": {
			"docType": "signature",
			"SigName": "%s"
		},
		"sort": [{
			"Timestamp": "desc"
		}],
		"limit": 1
	}`, sigName)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var signatures []*VirusSignature
	queryResponse, err := resultsIterator.Next()
	if err != nil {
		return nil, err
	}

	var signature VirusSignature
	err = json.Unmarshal(queryResponse.Value, &signature)
	if err != nil {
		return nil, err
	}

	signatures = append(signatures, &signature)

	return signatures, nil
}

func (t *VirusChaincode) GetAllSignatures(ctx contractapi.TransactionContextInterface) ([]*VirusSignature, error) {
	queryString := `{"selector":{"docType":"signature"}}`

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var signatures []*VirusSignature
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var signature VirusSignature
		err = json.Unmarshal(queryResponse.Value, &signature)
		if err != nil {
			return nil, err
		}

		signatures = append(signatures, &signature)
	}

	return signatures, nil
}

// Vote allows an organization to cast a vote if they haven't voted before
func (t *VirusChaincode) Vote(ctx contractapi.TransactionContextInterface, approve bool, signature_cid string) error {
	// Get the MSP ID of the invoking organization
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// Check if the organization has already voted
	voteAsBytes, err := ctx.GetStub().GetState(clientMSPID)
	if err != nil {
		return fmt.Errorf("failed to read vote from world state: %v", err)
	}
	if voteAsBytes != nil {
		return fmt.Errorf("organization %s has already voted", clientMSPID)
	}

	// Create a new vote
	vote := Vote{
    DocType: "vote",
		Org:     clientMSPID,
		Approve: approve,
		SignatureCID: signature_cid,
	}
	voteJSON, err := json.Marshal(vote)
	if err != nil {
		return fmt.Errorf("failed to marshal vote: %v", err)
	}

	// Store the vote in the world state with the organization's MSP ID as the key
	return ctx.GetStub().PutState(clientMSPID, voteJSON)
}

// GetVote retrieves the vote for an organization
// ListOrgsVoted lists all organizations that have voted along with their votes
func (t *VirusChaincode) ListVotes(ctx contractapi.TransactionContextInterface) ([]Vote, error) {
	queryString := `{"selector":{"docType":"vote"}}`

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to get query result: %v", err)
	}
	defer resultsIterator.Close()

	var orgVotes []Vote
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over results: %v", err)
		}

		var vote Vote
		err = json.Unmarshal(queryResponse.Value, &vote)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vote: %v", err)
		}

		orgVote := Vote{
			DocType:     vote.DocType,
			Org:     vote.Org,
			Approve: vote.Approve,
			SignatureCID: vote.SignatureCID,
		}
		orgVotes = append(orgVotes, orgVote)
	}

	return orgVotes, nil
}

func (t *VirusChaincode) CountApprovedVotesBySignatureCID(ctx contractapi.TransactionContextInterface, signatureCID string) (int, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"vote","approve":true,"signature_cid":"%s"}}`, signatureCID)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return 0, fmt.Errorf("failed to get query result: %v", err)
	}
	defer resultsIterator.Close()

	count := 0
	for resultsIterator.HasNext() {
		_, err := resultsIterator.Next()
		if err != nil {
			return 0, fmt.Errorf("failed to iterate over results: %v", err)
		}
		count++
	}

	return count, nil
}


func main() {
	virusChaincode := new(VirusChaincode)
	contractAPI, err := contractapi.NewChaincode(virusChaincode)
	if err != nil {
		log.Fatal("Error creating virus chaincode: ", err)
	}

	if err := contractAPI.Start(); err != nil {
		log.Fatal("Error starting virus chaincode: ", err)
	}
}
