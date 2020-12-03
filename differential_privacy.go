package checkpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/google/differential-privacy/go/dpagg"
	"github.com/google/differential-privacy/go/noise"
)

const address = "localhost"
const port = 8080

var consulConfigFields = []string{
	"acl_agent_token",
	"acl_datacenter",
	"acl_default_policy",
	"acl_down_policy",
	"acl_enable_key_list_policy",
	"acl_master_token",
	"acl_replication_token",
	"acl_token",
	"acl_ttl",
	"acl.enabled",
	"acl.down_policy",
	"acl.default_policy",
	"acl.enable_key_list_policy",
	"acl.enable_token_persistence",
	"acl.policy_ttl",
	"acl.role_ttl",
	"acl.token_ttl",
	"acl.enable_token_replication",
	"acl.msp_disable_bootstrap",
	"acl.tokens",
	"acl.tokens.master",
	"acl.tokens.agent_master",
	"acl.tokens.replication",
	"acl.tokens.agent",
	"acl.tokens.default",
	"acl.tokens.managed_service_provider",
	"acl.tokens.managed_service_provider.accessor_id",
	"acl.tokens.managed_service_provider.secret_id",
	"addresses.dns",
	"addresses.http",
	"addresses.https",
	"addresses.grpc",
	"advertise_addr",
	"advertise_addr_wan",
	"advertise_reconnect_timeout",
	"audit.enabled",
	"auto_config.enabled",
	"auto_config.intro_token",
	"auto_config.intro_token_file",
	"auto_config.dns_sans",
	"auto_config.ip_sans",
	"auto_config.server_addresses",
	"auto_config.authorization.enabled",
	"auto_config.authorization.static",
	"auto_config.authorization.static.allow_reuse",
	"auto_config.authorization.static.claim_mappings",
	"auto_config.authorization.static.claim_mappings.node",
	"auto_config.authorization.static.list_claim_mappings",
	"auto_config.authorization.static.list_claim_mappings.foo",
	"auto_config.authorization.static.bound_issuer",
	"auto_config.authorization.static.bound_audiences",
	"auto_config.authorization.static.claim_assertions",
	"auto_config.authorization.static.jwt_validation_pub_keys",
	"autopilot",
	"autopilot.cleanup_dead_servers",
	"autopilot.disable_upgrade_migration",
	"autopilot.last_contact_threshold",
	"autopilot.max_trailing_logs",
	"autopilot.min_quorum",
	"autopilot.redundancy_zone_tag",
	"autopilot.server_stabilization_time",
	"autopilot.upgrade_version_tag",
	"bind_addr",
	"bootstrap",
	"bootstrap_expect",
	"cache.entry_fetch_max_burst",
	"cache.entry_fetch_rate",
	"use_streaming_backend",
	"ca_file",
	"ca_path",
	"cert_file",
	"check.id",
	"check.name",
	"check.notes",
	"check.service_id",
	"check.token",
	"check.status",
	"check.args",
	"check.http",
	"check.header",
	"check.method",
	"check.body",
	"check.output_max_size",
	"check.tcp",
	"check.interval",
	"check.docker_container_id",
	"check.shell",
	"check.tls_skip_verify",
	"check.timeout",
	"check.ttl",
	"check.deregister_critical_service_after",
	"checks.id",
	"checks.name",
	"checks.notes",
	"checks.service_id",
	"checks.token",
	"checks.status",
	"checks.args",
	"checks.http",
	"checks.header",
	"checks.method",
	"checks.body",
	"checks.tcp",
	"checks.interval",
	"checks.output_max_size",
	"checks.docker_container_id",
	"checks.shell",
	"checks.tls_skip_verify",
	"checks.timeout",
	"checks.ttl",
	"checks.deregister_critical_service_after",
	"check_update_interval",
	"client_addr",
	"config_entries.bootstrap",
	"config_entries.bootstrap.kind",
	"config_entries.bootstrap.name",
	"config_entries.bootstrap.config",
	"auto_encrypt.tls",
	"auto_encrypt.dns_san",
	"auto_encrypt.ip_san",
	"auto_encrypt.allow_tls",
	"connect.ca_provider",
	"connect.ca_config",
	"connect.ca_config.rotation_period",
	"connect.ca_config.intermediate_cert_ttl",
	"connect.ca_config.leaf_cert_ttl",
	"connect.ca_config.csr_max_per_second",
	"connect.ca_config.csr_max_concurrent",
	"connect.enable_mesh_gateway_wan_federation",
	"connect.enabled",
	"gossip_lan.gossip_nodes",
	"gossip_lan.gossip_interval",
	"gossip_lan.retransmit_mult",
	"gossip_lan.suspicion_mult",
	"gossip_lan.probe_interval",
	"gossip_lan.probe_timeout",
	"gossip_wan",
	"gossip_wan.gossip_nodes",
	"gossip_wan.gossip_interval",
	"gossip_wan.retransmit_mult",
	"gossip_wan.suspicion_mult",
	"gossip_wan.probe_interval",
	"gossip_wan.probe_timeout",
	"data_dir",
	"datacenter",
	"default_query_time",
	"disable_anonymous_signature",
	"disable_coordinates",
	"disable_host_node_id",
	"disable_http_unprintable_char_filter",
	"disable_keyring_file",
	"disable_remote_exec",
	"disable_update_check",
	"discard_check_output",
	"discovery_max_stale",
	"domain",
	"alt_domain",
	"dns_config.allow_stale",
	"dns_config.a_record_limit",
	"dns_config.disable_compression",
	"dns_config.enable_truncate",
	"dns_config.max_stale",
	"dns_config.node_ttl",
	"dns_config.only_passing",
	"dns_config.recursor_timeout",
	"dns_config.service_ttl",
	"dns_config.udp_answer_limit",
	"dns_config.use_cache",
	"dns_config.cache_max_age",
	"dns_config.prefer_namespace",
	"enable_acl_replication",
	"enable_agent_tls_for_checks",
	"enable_central_service_config",
	"enable_debug",
	"enable_script_checks",
	"enable_local_script_checks",
	"enable_syslog",
	"encrypt",
	"encrypt_verify_incoming",
	"encrypt_verify_outgoing",
	"http_config.block_endpoints",
	"http_config.allow_write_http_from",
	"http_config.response_headers",
	"http_config.use_cache",
	"key_file",
	"leave_on_terminate",
	"limits.http_max_conns_per_client",
	"limits.https_handshake_timeout",
	"limits.rpc_handshake_timeout",
	"limits.rpc_rate",
	"limits.rpc_max_burst",
	"limits.rpc_max_conns_per_client",
	"limits.kv_max_value_size",
	"limits.txn_max_req_len",
	"log_level",
	"log_json",
	"max_query_time",
	"node_id",
	"node_meta",
	"node_name",
	"non_voting_server",
	"performance.leave_drain_time",
	"performance.raft_multiplier",
	"performance.rpc_hold_timeout",
	"pid_file",
	"ports.dns",
	"ports.http",
	"ports.https",
	"ports.server",
	"ports.grpc",
	"ports.sidecar_min_port",
	"ports.sidecar_max_port",
	"ports.expose_min_port",
	"ports.expose_max_port",
	"protocol",
	"primary_datacenter",
	"primary_gateways",
	"primary_gateways_interval",
	"raft_protocol",
	"raft_snapshot_threshold",
	"raft_snapshot_interval",
	"raft_trailing_logs",
	"read_replica",
	"reconnect_timeout",
	"reconnect_timeout_wan",
	"recursors",
	"rejoin_after_leave",
	"retry_interval",
	"retry_interval_wan",
	"retry_join",
	"retry_join_wan",
	"retry_max",
	"retry_max_wan",
	"rpc.enable_streaming",
	"segment",
	"segments.name",
	"segments.bind",
	"segments.port",
	"segments.rpc_listener",
	"segments.advertise",
	"serf_lan",
	"serf_wan",
	"server",
	"server_name",
	"service.id",
	"service.name",
	"service.meta",
	"service.tagged_addresses",
	"service.tagged_addresses.lan",
	"service.tagged_addresses.lan.address",
	"service.tagged_addresses.lan.port",
	"service.tagged_addresses.wan",
	"service.tagged_addresses.wan.address",
	"service.tagged_addresses.wan.port",
	"service.tags",
	"service.address",
	"service.token",
	"service.port",
	"service.weights",
	"service.weights.passing",
	"service.weights.warning",
	"service.enable_tag_override",
	"service.check.id",
	"service.check.name",
	"service.check.status",
	"service.check.notes",
	"service.check.args",
	"service.check.http",
	"service.check.header",
	"service.check.method",
	"service.check.body",
	"service.check.tcp",
	"service.check.interval",
	"service.check.output_max_size",
	"service.check.docker_container_id",
	"service.check.shell",
	"service.check.tls_skip_verify",
	"service.check.timeout",
	"service.check.ttl",
	"service.check.deregister_critical_service_after",
	"service.checks.id",
	"service.checks.name",
	"service.checks.notes",
	"service.checks.status",
	"service.checks.args",
	"service.checks.http",
	"service.checks.header",
	"service.checks.method",
	"service.checks.body",
	"service.checks.tcp",
	"service.checks.interval",
	"service.checks.output_max_size",
	"service.checks.docker_container_id",
	"service.checks.shell",
	"service.checks.tls_skip_verify",
	"service.checks.timeout",
	"service.checks.ttl",
	"service.checks.deregister_critical_service_after",
	"service.connect",
	"service.connect.native",
	"services.id",
	"services.name",
	"services.tags",
	"services.address",
	"services.token",
	"services.port",
	"services.enable_tag_override",
	"services.check",
	"services.check.id",
	"services.check.name",
	"services.check.status",
	"services.check.notes",
	"services.check.args",
	"services.check.http",
	"services.check.header",
	"services.check.method",
	"services.check.body",
	"services.check.tcp",
	"services.check.interval",
	"services.check.output_max_size",
	"services.check.docker_container_id",
	"services.check.shell",
	"services.check.tls_skip_verify",
	"services.check.timeout",
	"services.check.ttl",
	"services.check.deregister_critical_service_after",
	"services.connect",
	"services.connect.sidecar_service",
	"session_ttl_min",
	"skip_leave_on_interrupt",
	"start_join",
	"start_join_wan",
	"syslog_facility",
	"tagged_addresses",
	"telemetry.circonus_api_app",
	"telemetry.circonus_api_token",
	"telemetry.circonus_api_url",
	"telemetry.circonus_broker_id",
	"telemetry.circonus_broker_select_tag",
	"telemetry.circonus_check_display_name",
	"telemetry.circonus_check_force_metric_activation",
	"telemetry.circonus_check_id",
	"telemetry.circonus_check_instance_id",
	"telemetry.circonus_check_search_tag",
	"telemetry.circonus_check_tags",
	"telemetry.circonus_submission_interval",
	"telemetry.circonus_submission_url",
	"telemetry.disable_hostname",
	"telemetry.dogstatsd_addr",
	"telemetry.dogstatsd_tags",
	"telemetry.filter_default",
	"telemetry.prefix_filter",
	"telemetry.metrics_prefix",
	"telemetry.prometheus_retention_time",
	"telemetry.statsd_address",
	"telemetry.statsite_address",
	"telemetry.disable_compat_1.9",
	"tls_cipher_suites",
	"tls_min_version",
	"tls_prefer_server_cipher_suites",
	"translate_wan_addrs",
	"ui_config.enabled",
	"ui_config.dir",
	"ui_config.content_path",
	"ui_config.metrics_provider",
	"ui_config.metrics_provider_files",
	"ui_config.metrics_provider_options_json",
	"ui_config.metrics_proxy.base_url",
	"ui_config.metrics_proxy.add_headers",
	"ui_config.metrics_proxy.add_headers.name",
	"ui_config.metrics_proxy.add_headers.value",
	"ui_config.metrics_proxy.path_allowlist",
	"ui_config.dashboard_url_templates",
	"unix_sockets",
	"unix_sockets.group",
	"unix_sockets.mode",
	"unix_sockets.user",
	"verify_incoming",
	"verify_incoming_https",
	"verify_incoming_rpc",
	"verify_outgoing",
	"verify_server_hostname",
	"watches.type",
	"watches.datacenter",
	"watches.key",
	"watches.handler",
}

// DPClient stores our measurements locally
type DPClient struct {
	store  *store
	errors []error
}

// NewDPClient returns a client with all of its dpaggs initialized with default options.
func NewDPClient() DPClient {
	// Initialize agentsSum
	agentsSumOpts := &dpagg.BoundedSumInt64Options{
		Epsilon:                  epsilon,
		MaxPartitionsContributed: 1,
		Lower:                    0,
		Upper:                    100000,
		Noise:                    noise.Laplace(),
	}
	agentsSum := dpagg.NewBoundedSumInt64(agentsSumOpts)

	// Initialize agentsCount
	agentsCount := make(map[string]*dpagg.Count)
	agentsCountOpts := &dpagg.CountOptions{
		Epsilon:                  epsilon,
		MaxPartitionsContributed: 2,
		Noise:                    noise.Laplace(),
	}
	agentsCount["0-10"] = dpagg.NewCount(agentsCountOpts)
	agentsCount["11-100"] = dpagg.NewCount(agentsCountOpts)
	agentsCount["101-1000"] = dpagg.NewCount(agentsCountOpts)
	agentsCount["1001-10000"] = dpagg.NewCount(agentsCountOpts)
	agentsCount["10000+"] = dpagg.NewCount(agentsCountOpts)

	// Initialize config
	config := make(map[string]*dpagg.Count)
	configCountOpts := &dpagg.CountOptions{
		Epsilon:                  epsilon,
		MaxPartitionsContributed: 2,
	}
	for _, v := range consulConfigFields {
		config[v] = dpagg.NewCount(configCountOpts)
	}

	store := &store{
		agentSum:    agentsSum,
		agentCount:  agentsCount,
		configCount: config,
	}

	return DPClient{
		store: store,
	}
}

// Write takes a simulated cluster size and a simulated map of enabled agent config and applies their values to our
//  local client's differentially private stores
func (c *DPClient) Write(clusterSize int64, config map[string]bool) error {
	c.store.agentSum.Add(clusterSize)
	c.store.agentSumActual = clusterSize
	c.store.bucketConfigCount(config)
	if err := c.store.bucketAgentsCount(clusterSize); err != nil {
		return err
	}
	return nil
}

// Submit sends the DPClient's data to the aggregating server.
func (c *DPClient) Submit() error {
	address := fmt.Sprintf("http://%s:%d/submit", address, port)

	// Encode AgentsSum
	agents, err := c.store.agentSum.GobEncode()
	if err != nil {
		err = fmt.Errorf("error while encoding agent sum: %s", err)
		c.errors = append(c.errors, err)
		return err
	}

	// Encode AgentsCount
	agentsCount := make(map[string][]byte)
	for k, v := range c.store.agentCount {
		a, err := v.GobEncode()
		if err != nil {
			err = fmt.Errorf("error while encoding agent count %s: %s", k, err)
			c.errors = append(c.errors, err)
			return err
		}
		agentsCount[k] = a
	}

	// Encode ConfigCount
	configCount := make(map[string][]byte)
	for k, v := range c.store.configCount {
		a, err := v.GobEncode()
		if err != nil {
			err = fmt.Errorf("error while encoding config count %s: %s", k, err)
			c.errors = append(c.errors, err)
			return err
		}
		configCount[k] = a
	}

	s := submitBody{
		AgentsSum:    agents,
		AgentsCount:  agentsCount,
		AgentsActual: c.store.agentSumActual,
		ConfigCount:  configCount,
	}

	body, err := json.Marshal(s)
	if err != nil {
		c.errors = append(c.errors, err)
		return err
	}

	req, err := http.NewRequest("POST", address, bytes.NewBuffer(body))
	if err != nil {
		c.errors = append(c.errors, err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		c.errors = append(c.errors, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status code %d while submitting to server: %s",
			resp.StatusCode, err)
		c.errors = append(c.errors, err)
		return err
	}

	return nil
}

// Flush FIXME: is a noop. We should eventually use it to reset all of the dpaggs so we can encode them again
func (c *DPClient) Flush() {
	return
}

func (s *store) bucketAgentsCount(clusterSize int64) error {
	switch {
	case 0 <= clusterSize && clusterSize <= 10:
		s.agentCount["0-10"].Increment()
	case 11 <= clusterSize && clusterSize <= 100:
		s.agentCount["11-100"].Increment()
	case 101 <= clusterSize && clusterSize <= 1000:
		s.agentCount["101-1000"].Increment()
	case 1001 <= clusterSize && clusterSize <= 10000:
		s.agentCount["1001-10000"].Increment()
	case clusterSize > 10000:
		s.agentCount["10000+"].Increment()
	default:
		return fmt.Errorf("invalid clustersize '%d'. unable to bucket agent count", clusterSize)
	}
	return nil
}

func (s *store) bucketConfigCount(config map[string]bool) {
	for k, v := range config {
		if v {
			s.configCount[k].Increment()
		}
	}
}

func (s *store) bucketAgentsCountActual(clusterSize int64) error {
	switch {
	case 0 <= clusterSize && clusterSize <= 10:
		s.agentCountActual["0-10"]++
	case 11 <= clusterSize && clusterSize <= 100:
		s.agentCountActual["11-100"]++
	case 101 <= clusterSize && clusterSize <= 1000:
		s.agentCountActual["101-1000"]++
	case 1001 <= clusterSize && clusterSize <= 10000:
		s.agentCountActual["1001-10000"]++
	case clusterSize > 10000:
		s.agentCountActual["10000+"]++
	default:
		return fmt.Errorf("invalid clustersize '%d'. unable to bucket agent count actual",
			clusterSize)
	}
	return nil
}

// SimulateClientDonations simulates a client donating differentially privatized
// data
func SimulateClientDonations() error {
	client := NewDPClient()
	if err := client.Write(simulateClusterSize(), simulateConfig()); err != nil {
		return fmt.Errorf("error while writing simulated data: %s", err)
	}
	if err := client.Submit(); err != nil {
		return fmt.Errorf("error while submitting data: %s", err)
	}
	client.Flush()
	return nil
}

// simulateClusterSize returns a simulated number of agents in a cluster, weighting towards smaller cluster sizes rather
// than larger ones.
func simulateClusterSize() int64 {
	// weighted random number
	min := int64(0)
	max := int64(0)
	bucket := rand.Int63n(12)
	switch bucket {
	case 0, 1, 2, 3, 4:
		min = 0
		max = 10
	case 5, 6, 7, 8:
		min = 11
		max = 100
	case 9, 10:
		min = 101
		max = 1000
	case 11:
		min = 1001
		max = 10000
	default:
		min = 10001
		max = 100000
	}
	simulatedClusterSize := rand.Int63n(max-min+1) + min
	// Get abs value
	if simulatedClusterSize < 0 {
		simulatedClusterSize *= -1
	}
	return simulatedClusterSize
}

func simulateConfig() map[string]bool {
	config := make(map[string]bool)
	for _, f := range consulConfigFields {
		config[f] = false
		// 33% chance to be enabled
		if n := rand.Intn(100); n <= 33 {
			config[f] = true
		}
	}
	return config
}
