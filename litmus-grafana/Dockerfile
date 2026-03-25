ARG GRAFANA_VERSION=latest
FROM grafana/grafana:${GRAFANA_VERSION}

# Community plugins
RUN grafana-cli plugins install gapit-htmlgraphics-panel && \
    grafana-cli plugins install yesoreyeram-infinity-datasource && \
    grafana-cli plugins install volkovlabs-variable-panel && \
    grafana-cli plugins install marcusolsson-static-datasource && \
    grafana-cli plugins install nline-plotlyjs-panel

# Provisioned datasource (configured via env vars at runtime)
COPY --chown=grafana:root litmus-grafana/provisioning/ /etc/grafana/provisioning/

# Signed plugin (private signature, localhost rootUrls)
COPY --chown=grafana:root dist/ /var/lib/grafana/plugins/litmus-edge-datasource/

ENV GF_SERVER_ROOT_URL=http://localhost:3000
