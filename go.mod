module github.com/dedis/odyssey

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/sessions v1.2.0
	github.com/gorilla/websocket v1.4.0
	github.com/minio/cli v1.20.0
	github.com/minio/minio-go/v6 v6.0.34
	github.com/pkg/errors v0.8.1 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/swaggo/http-swagger v0.0.0-20190614090009-c2865af9083e
	github.com/swaggo/swag v1.6.3
	github.com/urfave/cli v1.22.2
	go.dedis.ch/cothority/v3 v3.1.3
	go.dedis.ch/kyber/v3 v3.0.12
	go.dedis.ch/onet/v3 v3.1.0
	go.dedis.ch/protobuf v1.0.11
	go.etcd.io/bbolt v1.3.3
	golang.org/x/xerrors v0.0.0-20191011141410-1b5146add898
	google.golang.org/appengine v1.4.0
	gopkg.in/urfave/cli.v1 v1.20.0
)

replace go.dedis.ch/cothority/v3 => /Users/nkocher/GitHub/cothority

replace github.com/dedis/odyssey => /Users/nkocher/GitHub/odyssey
