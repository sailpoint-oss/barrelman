module github.com/sailpoint-oss/barrelman

go 1.25.0

replace github.com/LukasParke/navigator => ../navigator

require (
	github.com/dlclark/regexp2 v1.11.5
	github.com/LukasParke/navigator v0.0.0-00010101000000-000000000000
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
	github.com/tree-sitter/go-tree-sitter v0.25.0
	github.com/vmware-labs/yaml-jsonpath v0.3.2
	github.com/yuin/goldmark v1.7.16
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/LukasParke/tree-sitter-openapi v0.1.0 // indirect
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960 // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
	golang.org/x/text v0.14.0 // indirect
)
