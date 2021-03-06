module seldon-mlops-task

go 1.14

require (
	github.com/seldonio/seldon-core/operator v0.0.0-20210305115125-18a2c688413c
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v12.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.18.8
