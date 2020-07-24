module github.com/dedis/odyssey

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/sessions v1.2.0
	github.com/minio/minio-go/v6 v6.0.53
	github.com/stretchr/testify v1.5.1
	github.com/swaggo/http-swagger v0.0.0-20200308142732-58ac5e232fba
	github.com/swaggo/swag v1.6.5 // indirect
	github.com/urfave/cli v1.22.4
	go.dedis.ch/cothority/v3 v3.4.4
	go.dedis.ch/kyber/v3 v3.0.12
	go.dedis.ch/onet/v3 v3.2.4
	go.dedis.ch/protobuf v1.0.11
	go.etcd.io/bbolt v1.3.4
	golang.org/x/sys v0.0.0-20200523222454-059865788121
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
)

replace go.dedis.ch/cothority/v3 => ../cothority
