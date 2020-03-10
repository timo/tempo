local frigg = import '../../../operations/jsonnet/single-binary/tempo.libsonnet';
local load = import 'synthetic-load-generator/main.libsonnet';

load + frigg {
    _config +:: {
        namespace: 'default',
        pvc_size: '30Gi',
        pvc_storage_class: 'local-path',
        receivers: {
            jaeger: {
              protocols: {
                thrift_http: null
              }
            }
        }
    },

    local container = $.core.v1.container,
    local containerPort = $.core.v1.containerPort,
    frigg_container+::
        $.util.resourcesRequests('1', '500Mi') +
        container.withPortsMixin([
            containerPort.new('jaeger-http', 14268),
        ]),

    local ingress = $.extensions.v1beta1.ingress,
    ingress:
        ingress.new() +
        ingress.mixin.metadata
            .withName('ingress')
            .withAnnotations({
                'ingress.kubernetes.io/ssl-redirect': 'false'
            }) +
        ingress.mixin.spec.withRules(
            ingress.mixin.specType.rulesType.mixin.http.withPaths(
                ingress.mixin.spec.rulesType.mixin.httpType.pathsType.withPath('/') +
                ingress.mixin.specType.mixin.backend.withServiceName('frigg') +
                ingress.mixin.specType.mixin.backend.withServicePort(16686)
            ),
        ),
}
