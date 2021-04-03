package mydynamo

const DYNAMO_CLIENT string = "[Dynamo RPCClient]:"
const DYNAMO_SERVER string = "[Dynamo Server]:"

const LOAD_FROM_DIR int = 0
const LOAD_FROM_METAFILE int = 1

const USAGE_STRING string = "usage: ./run-server [config file]"

//Server constants
const ARG_COUNT int = 2
const CONFIG_FILE_INDEX int = 1

//configuration constant labels
const EX_USAGE int = 2
const EX_CONFIG int = 3
const MYDYNAMO string = "mydynamo"
const SERVER_PORT string = "starting_port"
const W_VALUE string = "w_value"
const R_VALUE string = "r_value"
const CLUSTER_SIZE string = "cluster_size"
