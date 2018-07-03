package global

import (
	"github.com/qjpcpu/apiGate/mod"
)

const (
	SESSION_ID   = "sessionid"
	COMM_USER_ID = "user_id"
)

var G_conf_file string
var G mod.Configure
var G_cache *mod.Cache
var G_exit bool = false
