module github.com/chartmuseum/helm-push

go 1.13

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/bugsnag/bugsnag-go v1.5.0 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/emicklei/go-restful v2.16.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f // indirect
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1 // indirect
	github.com/spf13/cobra v1.8.0
	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940 // indirect
	github.com/yvasiyarov/gorelic v0.0.6 // indirect
	helm.sh/helm/v3 v3.14.3
	k8s.io/helm v2.16.1+incompatible
)

replace github.com/docker/docker => github.com/docker/docker v0.0.0-20190731150326-928381b2215c
