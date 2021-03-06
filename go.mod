module seldon-mlops-task

go 1.14

require (
	github.com/golang/protobuf v1.4.2
	github.com/seldonio/seldon-core/operator v0.0.0-20210305115125-18a2c688413c
	google.golang.org/protobuf v1.25.0
	//github.com/golang/protobuf v1.4.3
	//google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v12.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.18.8
