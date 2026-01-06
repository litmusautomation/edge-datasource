FROM grafana/grafana:12.3.1
ENV GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=litmus-edge-datasource
COPY dist /var/lib/grafana/plugins/litmus-edge-datasource
