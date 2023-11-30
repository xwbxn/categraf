module flashcat.cloud/categraf

go 1.18

require (
	github.com/chai2010/winsvc v0.0.0-20200705094454-db7ec320025c
	github.com/gin-gonic/gin v1.9.1
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.4
	github.com/influxdata/line-protocol/v2 v2.2.1
	github.com/json-iterator/go v1.1.12
	github.com/koding/multiconfig v0.0.0-20171124222453-69c27309b2d7
	github.com/mattn/go-isatty v0.0.19
	github.com/matttproud/golang_protobuf_extensions v1.0.4
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/client_model v0.3.0
	github.com/prometheus/common v0.39.0
	github.com/prometheus/prometheus v0.37.0
	github.com/shirou/gopsutil/v3 v3.22.5
	github.com/stretchr/testify v1.8.3
	github.com/toolkits/pkg v1.3.0
	golang.org/x/net v0.10.0
	golang.org/x/sys v0.8.0
	golang.org/x/text v0.9.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.22 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.1 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/digitalocean/go-smbios v0.0.0-20180907143718-390a4f403a8e // indirect
	github.com/frankban/quicktest v1.14.3 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/godbus/dbus/v5 v5.0.4 // indirect
	github.com/jaypipes/pcidb v1.0.0 // indirect
	github.com/jpillora/overseer v1.1.6 // indirect
	github.com/jpillora/s3 v1.1.4 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/smartystreets/assertions v1.1.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	golang.org/x/arch v0.3.0 // indirect
	howett.net/plist v1.0.0 // indirect
)

require (
	github.com/StackExchange/wmi v1.2.1
	github.com/aws/aws-sdk-go-v2 v1.17.4
	github.com/aws/aws-sdk-go-v2/config v1.18.12
	github.com/aws/aws-sdk-go-v2/credentials v1.13.12
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.3
	github.com/coreos/go-systemd/v22 v22.3.2
	github.com/jaypipes/ghw v0.12.0
	github.com/xwbxn/overseer v0.0.2
	github.com/yumaojun03/dmidecode v0.1.4
)

replace gopkg.in/yaml.v2 => github.com/rfratto/go-yaml v0.0.0-20211119180816-77389c3526dc

require (
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.0 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gopacket v1.1.19
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.9
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/ugorji/go/codec v1.2.11
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.uber.org/automaxprocs v1.5.1 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/googleapis/google-cloud-go/storage => cloud.google.com/go/storage v1.30.1
	github.com/prometheus/client_golang => github.com/flashcatcloud/client_golang v1.12.2-0.20220704074148-3b31f0c90903
	go.opentelemetry.io/collector => github.com/open-telemetry/opentelemetry-collector v0.54.0
)
