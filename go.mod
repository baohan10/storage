module agent

require (
	utils v0.0.0-00010101000000-000000000000
	ec v0.0.0-00010101000000-000000000000
	github.com/ipfs/go-ipfs-api v0.5.0
	google.golang.org/grpc v1.53.0
	proto v0.0.0
	rabbitmq v0.0.0-00010101000000-000000000000
)

replace proto => ../proto

replace rabbitmq => ../rabbitmq

replace utils => ../utils

replace ec => ../ec

require (
	github.com/beevik/etree v1.1.0 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/crackcomm/go-gitignore v0.0.0-20170627025303-887ab5e44cc3 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/go-ping/ping v1.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/ipfs/go-cid v0.3.2 // indirect
	github.com/ipfs/go-ipfs-files v0.0.9 // indirect
	github.com/klauspost/cpuid/v2 v2.2.1 // indirect
	github.com/klauspost/reedsolomon v1.11.7 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-flow-metrics v0.1.0 // indirect
	github.com/libp2p/go-libp2p v0.24.2 // indirect
	github.com/libp2p/go-openssl v0.1.0 // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.8.0 // indirect
	github.com/multiformats/go-multibase v0.1.1 // indirect
	github.com/multiformats/go-multicodec v0.7.0 // indirect
	github.com/multiformats/go-multihash v0.2.1 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/spacemonkeygo/spacelog v0.0.0-20180420211403-2296661a0572 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/streadway/amqp v1.0.0 // indirect
	github.com/whyrusleeping/tar-utils v0.0.0-20180509141711-8c6c8ba81d5c // indirect
	golang.org/x/crypto v0.3.0 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	lukechampine.com/blake3 v1.1.7 // indirect
)

go 1.18
