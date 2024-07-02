package constants

const (
	BLOCKCHAIN_NAME           = "Evochain"
	HEX_PREFIX                = "0x"
	SUCCESS                   = "success"
	FAILED                    = "failed"
	PENDING                   = "pending"
	MINING_DIFFICULTY         = 5
	MINING_REWARD             = 1200 * DECIMAL
	CURRENCY_NAME             = "evo"
	DECIMAL                   = 100
	BLOCKCHAIN_ADDRESS        = "Evochain_Faucet"
	BLOCKCHAIN_DB_PATH        = "5000/evodb"
	BLOCKCHAIN_KEY            = "blockchain_key"
	ADDRESS_PREFIX            = "evochain"
	TXN_VERIFICATION_SUCCESS  = "verification_success"
	TXN_VERIFICATION_FAILURE  = "verification_failure"
	BLOCKCHAIN_STATUS         = "RUNNING"
	PEER_BROADCAST_PAUSE_TIME = 1  // In seconds
	PEER_PING_PAUSE_TIME      = 60 // In seconds
	TXN_BROADCAST_PAUSE_TIME  = 1  // In seconds
	FETCH_LAST_N_BLOCKS       = 50
	CONSENSUS_PAUSE_TIME      = 10 // In seconds
)
