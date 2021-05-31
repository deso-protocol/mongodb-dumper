package mongodb

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/bitclout/core/lib"
	"log"
	"math/big"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/fatih/structs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// This file contains all sync functions associated with badgerDB and mongoDB

type SyncingService struct {
	// DB Holds a pointer to the global badgerDB database
	DB *badger.DB
	// SyncDBURI holds a string URI path for connecting to the running mongoDB server
	SyncDBURI string
	// mongoDBName holds a string dictating what database within the mongoDB client
	// to use for storing key/value pairs
	mongoDBName string
	// mongoCollectionName holds a string dictating which collection within
	// the mongoDBName specified database to use for storing key/value pairs
	mongoCollectionName string
	// mongoClient is a pointer to the mongo.Client object used for interfacing
	// with the mongo server dictated by SyncDBURI
	mongoClient *mongo.Client
}

// Initializes and returns a new SyncingService Structure with a nil mongo client
func NewSyncingService(db *badger.DB, syncDBURI string, mongoDBName string, mongoCollectionName string) *SyncingService {
	return &SyncingService{
		DB:                  db,
		SyncDBURI:           syncDBURI,
		mongoDBName:         mongoDBName,
		mongoCollectionName: mongoCollectionName,
		mongoClient:         nil,
	}
}

// This file contains all sync functions associated with badgerDB and mongoDB

// Establishes and returns a MongoDB client with associated URI MongoDbURI
func (syncSrv *SyncingService) ConnectToMongo() {
	// Establish MongoDB client options and create client
	clientOptions := options.Client().ApplyURI(syncSrv.SyncDBURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Failed establishing a connection with MongoDB: %v", err)
	}

	// Check MongoDB Connection and ensure data transmission
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB:  %v", err)
	}

	fmt.Println("Successfully Connected to MongoDB.")
	syncSrv.mongoClient = client
}

// Disconnects from MongoDB Client client
func (syncSrv *SyncingService) DisconnectFromMongo() {
	syncSrv.mongoClient.Disconnect(context.Background())
}

// Takes a map[string] interface {} and converts values into
// easier to read formats. Additionally adds a "Time" field with
// the current time for update purposes.
func SimplifyMap(docMap *map[string]interface{}) {
	// Go through all keys in map
	for key, val := range *docMap {
		if val == nil || (*docMap)[key] == nil {
			continue
		}

		switch val.(type) {
		case map[string]interface{}:
			subMap := (*docMap)[key].(map[string]interface{})
			SimplifyMap(&subMap)
			(*docMap)[key] = subMap
		case lib.UtxoType:
			(*docMap)[key] = (*docMap)[key].(lib.UtxoType).String()
		case lib.BlockHash:
			BH := (*docMap)[key].(lib.BlockHash)
			(*docMap)[key] = (&BH).String()
		case *lib.BlockHash:
			if (*docMap)[key].(*lib.BlockHash) == nil {
				(*docMap)[key] = nil          
			} else {
				(*docMap)[key] = (*docMap)[key].(*lib.BlockHash).String()
			}
		case *lib.BlockNode:
			prnt := (*docMap)[key].(*lib.BlockNode)
			delete((*docMap), key)
			if prnt != nil {
				(*docMap)["ParentHash"] = prnt.Hash.String()
			} else {
				(*docMap)["ParentHash"] = nil
			}
		case *big.Int:
			if (*docMap)[key].(*big.Int) == nil {
				(*docMap)[key] = nil          
			} else {
				(*docMap)[key] = (*docMap)[key].(*big.Int).String()
			}
		case lib.StakeEntry:
			delete(*docMap, key)
		}

		// ONLY USE KEY SWITCH FOR CHANGING PRIMITIVE TYPES
		// e.g. uint64, []byte
		switch key {
		case "PKID":
			(*docMap)[key] = lib.PkToStringBoth((*docMap)[key].([]byte))
		case "HODLerPKID":
			pkp := val.(*lib.PKID)
			pkbytes := make([]byte, len(*pkp))
			for i, v := range pkp {
				pkbytes[i] = v
			}
			(*docMap)[key] = lib.PkToStringBoth(pkbytes)
		case "PosterPublicKey", "FollowedPublicKey", "FollowedPKID", "FollowerPKID",
			"FollowedPublicKeys", "HoldingPublicKey", "PublicKey", "SenderPublicKey",
			"RecipientPublicKey":
			(*docMap)[key] = lib.PkToStringBoth((*docMap)[key].([]byte))
		case "CreatorPKID":
			pkp := val.(*lib.PKID)
			pkbytes := make([]byte, len(*pkp))
			for i, v := range pkp {
				pkbytes[i] = v
			}
			(*docMap)[key] = lib.PkToStringBoth(pkbytes)
		case "ParentStakeID":
			(*docMap)[key] = hex.EncodeToString((*docMap)[key].([]byte))
		case "Body":
			(*docMap)[key] = string((*docMap)[key].([]byte))
		case "Header":
			subm := (*docMap)[key].(map[string]interface{})
			if _, ok := subm["PrevBlockHash"].(string); !ok {
				subm["PrevBlockHash"] = subm["PrevBlockHash"].(*lib.BlockHash).String()
			}
			if _, ok := subm["TransactionMerkleRoot"].(string); !ok {
				subm["TransactionMerkleRoot"] = subm["TransactionMerkleRoot"].(*lib.BlockHash).String()
			}
			(*docMap)[key] = subm
		case "Txns":
			elems := (*docMap)[key].([]interface{})
			(*docMap)[key] = elems
		case "EncryptedText":
			(*docMap)[key] = string((*docMap)[key].([]byte))
		case "Username":
			(*docMap)[key] = string((*docMap)[key].([]byte))
		case "Description":
			(*docMap)[key] = string((*docMap)[key].([]byte))
		case "ProfilePic":
			(*docMap)[key] = string((*docMap)[key].([]byte))
		}
	}
	(*docMap)["Time"] = time.Now().String()
}

// Takes a badgerDB iterator pointer and returns its key's
// value formatted as a JSON
func BadgerItrToJSON(itr *badger.Iterator) []byte {
	key := itr.Item().Key() // Get key, val, prefix from itr
	prefix := key[0]

	// Debug setting
	/*if prefix != 0 {
		return nil
	}*/

	val, err := itr.Item().ValueCopy(nil)
	if err != nil {
		return nil
	}

	switch prefix {
	case 0: // _PrefixBlockHashToBlock
		blockRet := lib.NewMessage(lib.MsgTypeBlock).(*lib.MsgBitCloutBlock)
		var blockHash lib.BlockHash

		err = blockRet.FromBytes(val)
		if err != nil {
			return nil
		}
		copy(blockHash[:], key[1:])

		docMap := structs.Map(*blockRet)
		docMap["BlockHash"] = blockHash
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "A bitclout block and its corresponding blockhash."
		docMap["BadgerKeyPrefix"] = "_PrefixBlockHashToBlock:0"

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 1: // _PrefixHeightHashToNodeInfo
		BN, err := lib.DeserializeBlockNode(val)
		if err != nil {
			return nil
		}

		docMap := structs.Map(BN) // Convert to map
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "A block node in the bitclout blockchain graph."
		docMap["BadgerKeyPrefix"] = "_PrefixHeightHashToNodeInfo:1"

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 2: //_PrefixBitcoinHeightHashToNodeInfo
		BN, err := lib.DeserializeBlockNode(val)
		if err != nil {
			return nil
		}

		docMap := structs.Map(BN) // Convert to map
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "A block node in the bitcoin blockchain graph."
		docMap["BadgerKeyPrefix"] = "_PrefixBitcoinHeightHashToNodeInfo:2"

		docJSON, _ := json.Marshal(docMap)
		return docJSON
	case 3: //_KeyBestBitCloutBlockHash
		var ret lib.BlockHash
		_, err := itr.Item().ValueCopy(ret[:])
		if err != nil {
			return nil
		}

		docMap := map[string]interface{}{
			"Hash":            ret.String(),
			"MongoMeta":       "The hash of the front of the best BitClout Chain.",
			"BadgerKeyPrefix": "_KeyBestBitCloutBlockHash:3",
			"Time":            time.Now().String(),
		}

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 4: //_KeyBestBitcoinHeaderHash
		var ret lib.BlockHash
		_, err := itr.Item().ValueCopy(ret[:])
		if err != nil {
			return nil
		}

		docMap := map[string]interface{}{
			"Hash":            ret.String(),
			"MongoMeta":       "The hash of the front of the best Bitcoin Chain.",
			"BadgerKeyPrefix": "_KeyBestBitcoinHeaderHash:4",
			"Time":            time.Now().String(),
		}

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 5: //_PrefixUtxoKeyToUtxoEntry
		var ret lib.UtxoEntry
		dec := gob.NewDecoder(bytes.NewReader(val))
		err = dec.Decode(&ret)
		if err != nil {
			return nil
		}

		docMap := structs.Map(ret)
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "A UTXO Entry."
		docMap["BadgerKeyPrefix"] = "_PrefixUtxoKeyToUtxoEntry:5"

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 6: //_PrefixPositionToUtxoKey
		var ret lib.UtxoKey
		dec := gob.NewDecoder(bytes.NewReader(val))
		err = dec.Decode(&ret)
		if err != nil {
			return nil
		}

		docMap := structs.Map(ret)
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "A UTXO Key."
		docMap["BadgerKeyPrefix"] = "_PrefixPositionToUtxoKey:6"

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 7: //_PrefixPubKeyUtxoKey
		var ret lib.UtxoKey
		var newHash lib.BlockHash
		pubKey := key[1:34]
		copy(newHash[:], key[34:66])
		index := binary.BigEndian.Uint32(key[66:])

		ret.TxID = newHash
		ret.Index = index

		docMap := structs.Map(ret)
		docMap["PublicKey"] = pubKey
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "Public key and UTXO key."
		docMap["BadgerKeyPrefix"] = "_PrefixPubKeyUtxoKey:7"

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 8: //_KeyUtxoNumEntries
		numEntries := lib.DecodeUint64(val)

		docMap := map[string]interface{}{
			"UTXOs":           numEntries,
			"MongoMeta":       "The number of utxo entries in the database.",
			"BadgerKeyPrefix": "_KeyUtxoNumEntries:8",
			"Time":            time.Now().String(),
		}

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 9: //_PrefixBlockHashToUtxoOperations
		docMap := map[string]interface{}{
			"BadgerKeyPrefix": "_PrefixBlockHashToUtxoOperations:9",
			"Time":            time.Now().String(),
		}

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 10: //_KeyNanosPurchased
		var nanosPurchased uint64
		nanosPurchased = lib.DecodeUint64(val)

		docMap := map[string]interface{}{
			"Nanos":           nanosPurchased,
			"MongoMeta":       "The number of nanos purchased thus far.",
			"BadgerKeyPrefix": "_KeyNanosPurchased:10",
			"Time":            time.Now().String(),
		}

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 11: //_PrefixBitcoinBurnTxIDs
		var newHash lib.BlockHash
		copy(newHash[:], key[1:])

		docMap := map[string]interface{}{
			"TxID":            newHash,
			"MongoMeta":       "A processed bitcoin transaction.",
			"BadgerKeyPrefix": "_PrefixBitcoinBurnTxIDs:11",
			"Time":            time.Now().String(),
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 12: //_PrefixPublicKeyTimestampToPrivateMessage
		var ret lib.MessageEntry
		dec := gob.NewDecoder(bytes.NewReader(val))
		err = dec.Decode(&ret)
		if err != nil {
			return nil
		}

		docMap := structs.Map(ret)
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "An encrypted message between two users."
		docMap["BadgerKeyPrefix"] = "_PrefixPublicKeyTimestampToPrivateMessage:12"

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 13: //_KeyAccountData
		docMap := map[string]interface{}{
			"BadgerKeyPrefix": "_KeyAccountData:13",
			"Time":            time.Now().String(),
		}

		docJSON, _ := json.Marshal(docMap)
		return docJSON

		// TODO: Fix with proper data in the future
		//var accountData AccountData
		//_ := json.Unmarshal(val, &accountData)

	case 14: //_KeyTransactionIndexTip
		var ret *lib.BlockHash
		_, err = itr.Item().ValueCopy(ret[:])
		if err != nil {
			return nil
		}

		docMap := map[string]interface{}{
			"Hash": ret.String(),
			"MongoMeta": "The transaction index supports the block explorer and is only created when a node is run with --txindex." +
				"It uses its own separate blockchain data structure to create the index, and this is the tip of that blockchain.",
			"BadgerKeyPrefix": "_KeyTransactionIndexTip:14",
			"Time":            time.Now().String(),
		}

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 15: // _PrefixTransactionIDToMetadata
		var ret lib.TransactionMetadata
		dec := gob.NewDecoder(bytes.NewReader(val))
		err = dec.Decode(&ret)
		if err != nil {
			return nil
		}

		docMap := structs.Map(ret)
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "The transaction metadata for a particular transaction ID."
		docMap["BadgerKeyPrefix"] = "_PrefixTransactionIDToMetadata:15"

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 16: //_PrefixPublicKeyIndexToTransactionIDs
		// TODO: Fix with proper data in the future
		docMap := map[string]interface{}{
			"BadgerKeyPrefix": "_KeyAccountData:13",
			"Time":            time.Now().String(),
		}

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 17: // _PrefixPostHashToPostEntry
		dec := gob.NewDecoder(bytes.NewReader(val))
		var PE lib.PostEntry
		err = dec.Decode(&PE)
		if err != nil {
			return nil
		}

		docMap := structs.Map(PE) // Convert to map
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "A User's Post or Subcomment."
		docMap["BadgerKeyPrefix"] = "_PrefixPostHashToPostEntry:17"

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 18: //_PrefixPosterPublicKeyPostHash
		var newHash lib.BlockHash
		copy(newHash[:], key[34:])

		docMap := map[string]interface{}{
			"PublicKey":       key[1:34],
			"PostHash":        newHash,
			"MongoMeta":       "An association between a PostHash and its corresponding public key.",
			"BadgerKeyPrefix": "_PrefixPosterPublicKeyPostHash:18",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 19: // _PrefixTstampNanosPostHash
		var newHash lib.BlockHash
		copy(newHash[:], key[9:])

		docMap := map[string]interface{}{
			"TstampNanos":     lib.DecodeUint64(key[1:9]),
			"PostHash":        newHash,
			"MongoMeta":       "An association between a PostHash and its corresponding time stamp (in nanos).",
			"BadgerKeyPrefix": "_PrefixTstampNanosPostHash:19",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 20: // _PrefixCreatorBpsPostHash
		var newHash lib.BlockHash
		copy(newHash[:], key[9:])

		docMap := map[string]interface{}{
			"TstampNanos":     lib.DecodeUint64(key[1:9]),
			"PostHash":        newHash,
			"MongoMeta":       "An association between a PostHash and its corresponding creator basis points founder reward.",
			"BadgerKeyPrefix": "_PrefixCreatorBpsPostHash:20",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 21: // _PrefixMultipleBpsPostHash
		var newHash lib.BlockHash
		copy(newHash[:], key[9:])

		docMap := map[string]interface{}{
			"TstampNanos":     lib.DecodeUint64(key[1:9]),
			"PostHash":        newHash,
			"MongoMeta":       "An association between a PostHash and its multiplier basis points.",
			"BadgerKeyPrefix": "_PrefixMultipleBpsPostHash:21",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 22: // _PrefixCommentParentStakeIDToPostHash
		var newHash lib.BlockHash
		copy(newHash[:], key[42:])

		docMap := map[string]interface{}{
			"ParentStakeID":   key[1:34],
			"TstampNanos":     lib.DecodeUint64(key[34:42]),
			"PostHash":        newHash,
			"MongoMeta":       "An association between a comment PostHash and it's corresponding parent post stakeID.",
			"BadgerKeyPrefix": "_PrefixCommentParentStakeIDToPostHash:22",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 23: // _PrefixPKIDToProfileEntry
		dec := gob.NewDecoder(bytes.NewReader(val))
		var PE lib.ProfileEntry
		err = dec.Decode(&PE)
		if err != nil {
			return nil
		}

		docMap := structs.Map(PE) // Convert to map
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "A User's Profile."
		docMap["BadgerKeyPrefix"] = "_PrefixProfilePubKeyToProfileEntry:23"

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 24: // _PrefixProfileStakeToProfilePubKey
		docMap := map[string]interface{}{
			"Stake":           lib.DecodeUint64(key[1:9]),
			"PublicKey":       key[9:],
			"MongoMeta":       "Depricated.",
			"BadgerKeyPrefix": "_PrefixProfileStakeToProfilePubKey:24",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 25: // _PrefixProfileUsernameToPKID
		pubKey := val
		docMap := map[string]interface{}{
			"Username":        key[1:],
			"PKID":            pubKey,
			"MongoMeta":       "A user's username and their corresponding PKID.",
			"BadgerKeyPrefix": "_PrefixProfileUsernameToProfilePubKey:25",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 26: // _PrefixStakeIDTypeAmountStakeIDIndex
		docMap := map[string]interface{}{
			"StakeType":       "",
			"AmountNanos":     key[2:10],
			"StakeID":         key[10:],
			"MongoMeta":       "A stake ID and its corresponding stakes nanos.",
			"BadgerKeyPrefix": "_PrefixStakeIDTypeAmountStakeIDIndex:26",
		}
		if key[1] == 0 {
			docMap["StakeType"] = "Post"
		} else {
			docMap["StakeType"] = "Profile"
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 27: // _KeyUSDCentsPerBitcoinExchangeRate
		var exchange uint64
		exchange = lib.DecodeUint64(val)

		docMap := map[string]interface{}{
			"USDCentsPerBitcoin": exchange,
			"MongoMeta":          "The exchange rate in USD Cents for a bitcoin.",
			"BadgerKeyPrefix":    "_KeyUSDCentsPerBitcoinExchangeRate:27",
			"Time":               time.Now().String(),
		}

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 28: // _PrefixFollowerPKIDToFollowedPKID
		docMap := map[string]interface{}{
			"FollowerPKID":    key[1:34],
			"FollowedPKID":    key[34:],
			"MongoMeta":       "A user's PKID (follower) and the PKID of those they follow (followed).",
			"BadgerKeyPrefix": "_PrefixFollowerPubKeyToFollowedPubKey:28",
			"Time":            time.Now().String(),
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 29: // _PrefixFollowedPubKeyToFollowerPubKey
		docMap := map[string]interface{}{
			"FollowedPKID":    key[1:34],
			"FollowerPKID":    key[34:],
			"MongoMeta":       "A user's PKID (followed) and the PKID of those who follow them (follower).",
			"BadgerKeyPrefix": "_PrefixFollowedPubKeyToFollowerPubKey:29",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 30: // _PrefixLikerPubKeyToLikedPostHash
		var likedPost lib.BlockHash
		copy(likedPost[:], key[34:])
		docMap := map[string]interface{}{
			"PublicKey":       key[1:34],
			"LikedPostHash":   likedPost,
			"MongoMeta":       "A user's public key and the post hash of one of their liked posts.",
			"BadgerKeyPrefix": "_PrefixLikerPubKeyToLikedPostHash:30",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 31: // _PrefixLikedPostHashToLikerPubKey
		var likedPost lib.BlockHash
		copy(likedPost[:], key[1:34])
		docMap := map[string]interface{}{
			"PublicKey":       key[1:34],
			"LikedPostHash":   likedPost,
			"MongoMeta":       "A PostHash and a corresponding public key of someone who liked that post.",
			"BadgerKeyPrefix": "_PrefixLikedPostHashToLikerPubKey:31",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 32: // _PrefixCreatorBitCloutLockedNanosCreatorPKID
		docMap := map[string]interface{}{
			"PKID":                key[9:],
			"BitCloutLockedNanos": lib.DecodeUint64(key[1:9]),
			"MongoMeta":           "The amount of BitClout locked in a particular profile's PKID.",
			"BadgerKeyPrefix":     "_PrefixCreatorBitCloutLockedNanosCreatorPubKeyIIndex:32",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 33: // _PrefixHODLerPubKeyCreatorPubKeyToBalanceEntry
		dec := gob.NewDecoder(bytes.NewReader(val))
		var BE lib.BalanceEntry
		err = dec.Decode(&BE)
		if err != nil {
			return nil
		}

		docMap := structs.Map(BE)
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "A user's (HODLerPubKey) balance of held (CreatorPubKey)."
		docMap["BadgerKeyPrefix"] = "_PrefixHODLerPubKeyCreatorPubKeyToBalanceEntry:33"

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 34: // _PrefixCreatorPubKeyHODLerPubKeyToBalanceEntry
		dec := gob.NewDecoder(bytes.NewReader(val))
		var BE lib.BalanceEntry
		err = dec.Decode(&BE)
		if err != nil {
			return nil
		}

		docMap := structs.Map(BE)
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "A ceator's (CreatorPubKey) hodlers (HODLerPubKey) and their associated balances."
		docMap["BadgerKeyPrefix"] = "_PrefixCreatorPubKeyHODLerPubKeyToBalanceEntry:34"

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 35: // _PrefixPosterPublicKeyTimestampPostHash
		var ph lib.BlockHash
		copy(ph[:], key[42:])

		docMap := map[string]interface{}{
			"PublicKey":       key[1:34],
			"PostHash":        ph,
			"TStampNanos":     lib.DecodeUint64(key[34:42]),
			"MongoMeta":       "The PostHash of a post generated by a user's public key and the corresponding time in nanos.",
			"BadgerKeyPrefix": "_PrefixPosterPublicKeyTimestampPostHash:35",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 36: // _PrefixPublicKeyToPKID
		dec := gob.NewDecoder(bytes.NewReader(val))
		var PE lib.PKIDEntry
		err = dec.Decode(&PE)
		if err != nil {
			return nil
		}

		docMap := structs.Map(PE)
		pkp := docMap["PKID"].(*lib.PKID)
		pk := make([]byte, len(*pkp))
		for i, v := range *pkp {
			pk[i] = v
		}
		docMap["PKID"] = pk

		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "A mapping of a public key to it's corresponding PKID."
		docMap["BadgerKeyPrefix"] = "_PrefixPublicKeyToPKID:36"

		docJSON, _ := json.Marshal(docMap)
		return docJSON

	case 37: // _PrefixPKIDToPublicKey
		docMap := map[string]interface{}{
			"PublicKey":       val,
			"PKID":            key[1:34],
			"MongoMeta":       "A map of a PKID to it's corresponding public key.",
			"BadgerKeyPrefix": "_PrefixPKIDToPublicKey:37",
		}
		SimplifyMap(&docMap)

		docJSON, _ := json.Marshal(docMap)
		return docJSON
	case 39: //_PrefixReclouterPubKeyRecloutedPostHashToRecloutPostHash

		dec := gob.NewDecoder(bytes.NewReader(val))
		var RE lib.RecloutEntry
		err = dec.Decode(&RE)
		if err != nil {
			return nil
		}

		docMap := structs.Map(RE) // Convert to map
		SimplifyMap(&docMap)
		docMap["MongoMeta"] = "A user's public key and the post hash of one of the post they reclouted"
		docMap["BadgerKeyPrefix"] = "_PrefixReclouterPubKeyRecloutedPostHashToRecloutPostHash:39"

		docJSON, _ := json.Marshal(docMap)
		return docJSON
	case 40: // _KeyGlobalParams

		dec := gob.NewDecoder(bytes.NewReader(val))
		var GPE lib.GlobalParamsEntry
		err = dec.Decode(&GPE)
		if err != nil {
			return nil
		}
		docMap := structs.Map(GPE)
		docMap["MongoMeta"] = "Global Params Entry"
		docMap["BadgerKeyPrefix"] = "_KeyGlobalPArams"
		docJSON, _ := json.Marshal(docMap)
		return docJSON
	default:
		return nil
	}
}

// Starts syncing badgerDB data to mongoDB client
func (syncSrv *SyncingService) Start() {
	if syncSrv.mongoClient == nil {
		fmt.Println("Failed to start mongoDB Sync. Invalid Mongo Client." +
			"Check for proper mongoDB URI.")
		return
	}
	mongodb := syncSrv.mongoClient.Database(syncSrv.mongoDBName)
	MongoCollection := mongodb.Collection(syncSrv.mongoCollectionName)

	for {
		err := syncSrv.DB.View(func(txn *badger.Txn) error {
			itr := txn.NewIterator(badger.DefaultIteratorOptions)
			defer itr.Close()

			totalIterations := 0
			bulkWriteChunkSize := 1000 // Number of operations in a bulk write operation
			var ops []mongo.WriteModel

			// Here we iterate over all keys in BadgerDB. itr.Valid() is only
			// false if we've reached the end of BadgerDB.
			for itr.Seek(nil); itr.Valid(); itr.Next() {

				// Execute MongoDB Bulk Write
				if (totalIterations%bulkWriteChunkSize) == 0 && totalIterations != 0 {
					bulkOption := options.BulkWriteOptions{}
					bulkOption.SetOrdered(false) // Continues writes even if an error occurs
					_, er := MongoCollection.BulkWrite(context.Background(), ops, &bulkOption)
					ops = nil

					if er != nil {
						fmt.Println("Failed MongoDB bulk write")
					} else {
						fmt.Println("Completed MongoDB BulkWrite.")
					}

					totalIterations = 0 // Reset total iterations to prevent overflow
				}

				// Convert badger iterator to JSON
				docJSON := BadgerItrToJSON(itr)
				if docJSON == nil {
					continue
				}

				// Unmarshal JSON into BSON
				var docBSON map[string]interface{}
				err := json.Unmarshal(docJSON, &docBSON)
				if err != nil {
					continue
				}

				// Create and add operation
				op := mongo.NewUpdateOneModel()
				op.SetFilter(bson.M{"_id": string(itr.Item().Key())})
				op.SetUpdate(bson.M{"$set": docBSON})
				op.SetUpsert(true)
				ops = append(ops, op)
				totalIterations++
			}

			// Push remaining bulk operations
			if totalIterations != 0 {
				bulkOption := options.BulkWriteOptions{}
				bulkOption.SetOrdered(false)
				_, err := MongoCollection.BulkWrite(context.Background(), ops[0:totalIterations], &bulkOption)

				if err != nil {
					fmt.Println("Failed bulk write")
				}
			}

			// Wait a minute before conintuing to limit CPU utilization
			time.Sleep(60 * time.Second)

			return nil
		})
		if err != nil {
			fmt.Println("Ran into problem processing Mongo...")
		}
	}
}
