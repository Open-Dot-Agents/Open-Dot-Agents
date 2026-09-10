#!/usr/bin/env python3
"""Check semantic context and frozen source coverage."""
import collections
import json
import unittest
from native_coverage import ROOT, INVENTORY, build, classify, compiled_registry


class NativeCoverageTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.inventory = json.loads(INVENTORY.read_text())
        cls.registry = compiled_registry()
        cls.result = build(cls.inventory, cls.registry)

    def feature(self, vendor, context, name):
        return next(f for f in self.result['features'] if (f['vendor'], f['context'], f['native']) == (vendor, context, name))

    def test_original_rows_exactly_once(self):
        rows = [e['entry_id'] for f in self.result['features'] for e in f['evidence']]
        self.assertEqual(collections.Counter(rows), collections.Counter('source-entry-' + str(i) for i in range(1, 1680)))

    def test_plugin_discovery_is_separate_from_execution(self):
        for vendor, names in [('codex', ['plugins.<name>.enabled', 'marketplaces.<name>.source', 'marketplaces.<name>.ref']),
                              ('copilot', ['enabledPlugins', 'extraKnownMarketplaces'])]:
            for name in names:
                feature = self.feature(vendor, 'settings', name)
                self.assertEqual(feature['native_status'], 'native-discovery-only')
                self.assertEqual(feature['disposition'], 'artifact-field-mapping')
                self.assertEqual(feature['implemented_scopes'], ['project', 'user'])
                self.assertTrue(any('selection-mcp-reviewed-' in path for path in feature['native_evidence']))
                self.assertTrue(any('git-mcp-reviewed-' in path for path in feature['native_evidence']))
            fields = [f for f in self.result['native_artifacts'][vendor] if f['feature'].startswith('artifact:plugins:')]
            self.assertTrue(fields)
            for field in fields:
                self.assertEqual(field['native_status'], 'native-discovery-only')
                self.assertIn('Public HTTPS/SSH services', field['limitation'])
        self.assertEqual(self.feature('codex', 'settings', 'plugins.<name>.mcp_servers.<name>.enabled')['native_status'], 'unverified')

    def test_plugin_mcp_execution_keeps_native_conformance_limits(self):
        for vendor in ('codex', 'copilot'):
            fields = [f for f in self.result['native_artifacts'][vendor] if f['feature'] == 'artifact:plugin-component:stdio']
            self.assertEqual(len(fields), 2)
            for field in fields:
                self.assertEqual(field['native_status'], 'bounded-fixture-execution')
                self.assertEqual(field['disposition'], 'native-operation')
                self.assertIn('not full package-format conformance', field['limitation'])
                if vendor == 'copilot': self.assertIn('violate the standard', field['limitation'])
                self.assertEqual(len(field['evidence']), 4 if vendor == 'codex' else 3)

    def test_copilot_preferences_do_not_share_permission_or_state_context(self):
        memory = self.feature('copilot', 'settings', 'memory')
        self.assertEqual(memory['disposition'], 'validator-declared')
        self.assertEqual(memory['implemented_scopes'], ['user'])
        self.assertEqual(self.feature('copilot', 'saved-permissions', 'memory')['disposition'], 'external')
        trust = self.feature('copilot', 'settings', 'trustedFolders')
        self.assertEqual(trust['category'], 'external-authority-state')
        self.assertEqual(trust['disposition'], 'external')
        for parent in ('ide', 'subagents', 'tabs'):
            self.assertIn(parent, self.registry['copilot'])

    def test_copilot_user_settings_are_not_state(self):
        for name in ('autoUpdate', 'banner'):
            feature = self.feature('copilot', 'settings', name)
            self.assertEqual(feature['category'], 'native-configuration')
            self.assertEqual(feature['disposition'], 'validator-declared')
            self.assertEqual(feature['implemented_scopes'], ['user'])

    def test_preference_evidence_is_scoped_and_limited(self):
        for name in ('defaultMode', 'tabs.enabled', 'tabs.sort', 'tabs.hide', 'statusLine'):
            feature = self.feature('copilot', 'settings', name)
            self.assertEqual(feature['native_status'], 'bounded-fixture-execution')
            self.assertEqual(feature['implemented_scopes'], ['user'])
            self.assertEqual(feature['native_evidence'], ['WORKBENCH/evidence/native-draft2-debug/copilot-preferences-final.json'])
        self.assertIn('Autopilot execution', self.feature('copilot', 'settings', 'defaultMode')['limitations'][0])
        for name in ('inlineImages', 'inlineImageLiveWindow', 'notifications'):
            self.assertEqual(self.feature('copilot', 'settings', name)['native_status'], 'unverified')

    def test_subagent_mapping_distinguishes_account_prerequisites(self):
        for name in ('subagents.agents', 'subagents.agents.<name>', 'subagents.disabledSubagents'):
            feature = self.feature('copilot', 'settings', name)
            self.assertEqual(feature['native_status'], 'bounded-fixture-execution')
            self.assertEqual(feature['implemented_scopes'], ['user'])
        for name in ('subagents.maxDepth', 'subagents.maxConcurrency'):
            feature = self.feature('copilot', 'settings', name)
            self.assertEqual(feature['native_status'], 'native-prerequisite-required')
            self.assertIn('usage-based billing', feature['limitations'][0])
            self.assertIn('two nested agents', feature['limitations'][0])

    def test_schema_array_indexes_resolve_without_new_ids(self):
        for leaf in ('path', 'enabled'):
            feature = self.feature('codex', 'settings', 'skills.config.<name>.' + leaf)
            self.assertEqual(feature['disposition'], 'artifact-field-mapping')
            self.assertEqual(feature['native_status'], 'bounded-fixture-context')
            self.assertEqual(feature['implemented_scopes'], ['user'])
            self.assertTrue(any(e['source'] == 'codex-config' for e in feature['evidence']))

    def test_otel_exporter_aliases_keep_finite_schema_choices(self):
        for exporter in ('exporter', 'trace_exporter'):
            for field in ('endpoint', 'headers', 'protocol', 'tls.ca-certificate', 'tls.client-certificate', 'tls.client-private-key'):
                feature = self.feature('codex', 'settings', 'otel.' + exporter + '.<name>.' + field)
                choices = ['otlp-http'] if field == 'protocol' else ['otlp-grpc', 'otlp-http']
                self.assertEqual(feature['validation_paths'], ['otel.' + exporter + '.' + choice + '.' + field for choice in choices])
                self.assertEqual(feature['implemented_scopes'], ['user'])
                identity = field.startswith('tls.client-')
                self.assertEqual(feature['native_status'], 'unverified' if identity else 'bounded-fixture-execution')
                self.assertEqual(feature['disposition'], 'validator-declared' if identity else 'artifact-field-mapping')
                if identity:
                    self.assertEqual(len(feature['native_variant_limitations']), 1)
                    variant = feature['native_variant_limitations'][0]
                    self.assertEqual(variant['native_status'], 'native-exporter-failure')
                    self.assertEqual(variant['activation'], 'inactive')
                    self.assertIn('/otlp-http/', variant['feature'])
        self.assertNotIn('otel', [f['feature'] for f in self.registry['codex']['__roots__'] if f['scope'] == 'project'])
        self.assertEqual(self.feature('codex', 'settings', 'otel.metrics_exporter')['native_status'], 'unverified')

    def test_frontmatter_does_not_reuse_root_setting_validator(self):
        feature = self.feature('copilot', 'agent-frontmatter', 'model')
        self.assertEqual(feature['disposition'], 'validator-declared')
        self.assertEqual(feature['native_status'], 'unverified')
        self.assertTrue(feature['validation_source'].endswith('native_copilot_agent.go'))
        self.assertNotEqual(feature['id'], self.feature('copilot', 'settings', 'model')['id'])

    def test_managed_model_has_separate_authority(self):
        feature = self.feature('copilot', 'managed-policy', 'model')
        self.assertEqual(feature['disposition'], 'external')
        self.assertEqual(feature['scope'], ['managed'])

    def test_mcp_fields_are_not_commands(self):
        self.assertEqual(self.feature('copilot', 'mcp', 'command')['category'], 'portable-configuration')
        self.assertEqual(self.feature('copilot', 'mcp', 'headers')['category'], 'native-configuration')

    def test_mcp_fields_use_the_native_artifact_validator(self):
        for name in ('command', 'headers', 'tools'):
            feature = self.feature('copilot', 'mcp', name)
            self.assertTrue(feature['validation_source'].endswith('native_copilot_mcp.go'))
            self.assertEqual(feature['native_status'], 'bounded-fixture-execution')
        self.assertEqual(self.feature('copilot', 'mcp', 'timeout')['native_status'], 'unverified')

    def test_agent_local_mcp_has_separate_evidence(self):
        feature = self.feature('copilot', 'agent-frontmatter', 'mcp-servers')
        self.assertEqual(feature['native_status'], 'bounded-fixture-execution')
        self.assertTrue(feature['validation_source'].endswith('native_copilot_agent.go'))
        fields = [f for f in self.result['native_artifacts']['copilot'] if f['feature'] == 'artifact:agent:/mcp-servers']
        self.assertEqual({f['scope'] for f in fields}, {'project', 'user'})
        for field in fields:
            self.assertTrue(any('agent-mcp-final-' + field['scope'] in path for path in field['evidence']))
            self.assertTrue(any('agent-mcp-override-' + field['scope'] in path for path in field['evidence']))

    def test_hook_files_and_inline_settings_have_separate_evidence(self):
        feature = self.feature('copilot', 'settings', 'hooks')
        self.assertEqual(feature['native_status'], 'bounded-fixture-execution')
        self.assertTrue(feature['validation_source'].endswith('native_copilot_hooks.go'))
        self.assertTrue(all('inline' in path for path in feature['native_evidence']))
        self.assertEqual(self.feature('copilot', 'settings', 'disableAllHooks')['disposition'], 'security-evidence-required')
        fields = [f for f in self.result['native_artifacts']['copilot'] if f['feature'].startswith('artifact:hooks:/')]
        self.assertEqual({f['scope'] for f in fields}, {'project', 'user'})
        for field in fields:
            if field['feature'].endswith('/prompt'):
                self.assertEqual(field['native_status'], 'unverified')
            if field['feature'].endswith('/timeout'):
                self.assertTrue(any('timeout-alias-' + field['scope'] in p for p in field['evidence']))

    def test_codex_hooks_use_native_registry_and_scope_evidence(self):
        root = self.feature('codex', 'settings', 'hooks')
        self.assertEqual(root['native_status'], 'bounded-fixture-execution')
        self.assertTrue(all('inline-' in path for path in root['native_evidence']))
        for name in ('hooks.<name>', 'hooks.<name>[].hooks', 'hooks.<name>[].hooks[].commandWindows'):
            feature = self.feature('codex', 'settings', name)
            self.assertEqual(feature['disposition'], 'validator-declared')
            self.assertEqual(feature['native_status'], 'unverified')
            self.assertTrue(feature['validation_source'].endswith('native_codex_hooks.go'))

        for name in ('hooks.<name>[].hooks[].async', 'hooks.<name>[].hooks[].additionalContextLimit'):
            feature = self.feature('codex', 'settings', name)
            self.assertEqual(feature['disposition'], 'artifact-field-mapping')
            self.assertEqual(feature['native_status'], 'bounded-fixture-execution')
            self.assertTrue(any('codex-background-final-' in path for path in feature['native_evidence']))

    def test_codex_mcp_hooks_have_separate_evidence(self):
        fields = [f for f in self.result['native_artifacts']['codex']
                  if '/hooks/<mcp_tool>/' in f['feature']]
        self.assertEqual(len(fields), 12)
        for field in fields:
            self.assertEqual(field['native_status'], 'bounded-fixture-execution')
            self.assertTrue(any('codex-mcp-hooks-final-' in path and
                                field['scope'] in path for path in field['evidence']))
            if field['feature'].endswith('/timeout'):
                self.assertTrue(all('timeout-' in path for path in field['evidence']))
                self.assertIn('does not prove MCP tool termination', field['limitation'])
            if field['feature'].endswith('/input'):
                self.assertIn('Null input values are refused', field['limitation'])

    def test_codex_background_and_context_limits_have_execution_evidence(self):
        for setting, required in [('async', ['background-next', 'background-active', 'background-deny', 'background-timeout', 'background-detached-timeout', 'background-unsubscribe', 'background-shutdown', 'background-archive']),
                                  ('additionalContextLimit', ['spill', 'spill-tiny', 'unlimited', 'spill-default', 'spill-mixed', 'background-spill', 'background-unlimited'])]:
            fields = [f for f in self.result['native_artifacts']['codex']
                      if f['feature'].endswith('/hooks/<command>/' + setting)]
            self.assertEqual(len(fields), 2)
            for field in fields:
                self.assertEqual(field['native_status'], 'bounded-fixture-execution')
                for scenario in required:
                    self.assertTrue(any('codex-background-final-' + scenario + '-' + field['scope'] in path
                                        for path in field['evidence']))
                if setting == 'async':
                    self.assertIn('a detached child can outlive the timeout', field['limitation'])
                else:
                    self.assertIn('spill-write failures need separate tests', field['limitation'])

    def test_alias_does_not_inflate_count(self):
        feature = self.feature('codex', 'settings', 'agents.max_concurrent_threads_per_session')
        self.assertIn('agents.max_threads', {e['native_name'] for e in feature['evidence']})

    def test_identity_does_not_depend_on_implementation(self):
        other = build(self.inventory, {'codex': {}, 'copilot': {}})
        self.assertEqual([f['id'] for f in self.result['features']], [f['id'] for f in other['features']])
        self.assertFalse(self.result['complete'])
        self.assertFalse(self.result['counts_are_completion'])

    def test_classification_retains_existing_ids_and_source_assignments(self):
        baseline = json.loads((ROOT / 'WORKBENCH/evidence/native-draft2-debug/coverage-classification-baseline.json').read_text())
        current = {f['id']: sorted(e['entry_id'] for e in f['evidence']) for f in self.result['features']}
        self.assertEqual(current, baseline['identities'])
        self.assertEqual(self.result['source_sha256'], baseline['source_sha256'])

    def test_permission_values_and_selectors_are_linked_syntax(self):
        parent = self.feature('codex', 'settings', 'permissions.<name>.filesystem.<name>')
        for name in ('read', 'write', 'deny', ':root', ':minimal', ':workspace_roots', ':tmpdir', ':slash_tmp', '/absolute/path', '~/path'):
            feature = self.feature('codex', 'settings', name)
            self.assertEqual(feature['disposition'], 'syntax-reference')
            self.assertFalse(feature['counted_feature'])
            self.assertEqual(feature['describes'], [parent['id']])
            self.assertEqual(feature['native_status'], 'not-applicable')
            self.assertIsNone(feature['adapter'])
        self.assertEqual(parent['disposition'], 'security-evidence-required')
        self.assertEqual(parent['implemented_scopes'], [])
        self.assertEqual(parent['gate_scopes'], ['project', 'user'])

    def test_table_notation_and_empty_binding_do_not_inflate_counts(self):
        alias = self.feature('codex', 'settings', '[permissions.<name>.filesystem].<name>')
        parent = self.feature('codex', 'settings', 'permissions.<name>.filesystem.<name>')
        self.assertEqual(alias['record_kind'], 'alias')
        self.assertEqual(alias['describes'], [parent['id']])
        self.assertFalse(alias['counted_feature'])
        nested = self.feature('codex', 'settings', '[permissions.<name>.filesystem.<name>].<name>')
        self.assertTrue(nested['counted_feature'])
        self.assertEqual(nested['setting_path'], 'permissions.<name>.filesystem.<name>.<name>')
        self.assertEqual(nested['disposition'], 'security-evidence-required')
        unbind = self.feature('codex', 'settings', 'tui.keymap.<name>.<name> = []')
        self.assertEqual(unbind['describes'], [self.feature('codex', 'settings', 'tui.keymap.<name>.<name>')['id']])
        self.assertFalse(unbind['counted_feature'])

    def test_desktop_scope_and_platform_boundaries_have_explicit_dispositions(self):
        desktop = self.feature('codex', 'settings', 'desktop.custom_file_handlers.<name>.command')
        self.assertEqual(desktop['scope'], ['user'])
        self.assertEqual(desktop['disposition'], 'outside-milestone')
        self.assertEqual(desktop['milestone_scope'], 'outside-linux-cli')
        windows = self.feature('codex', 'settings', 'computer_use.windows.aumids')
        self.assertEqual(windows['disposition'], 'outside-milestone')
        managed = self.feature('codex', 'requirements', 'computer_use.windows.aumids')
        self.assertEqual(managed['disposition'], 'external')
        state = self.feature('codex', 'settings', 'windows_wsl_setup_acknowledged')
        self.assertEqual(state['category'], 'external-authority-state')
        self.assertEqual(state['disposition'], 'external')

    def test_counts_separate_records_features_and_milestone(self):
        records = self.result['features']
        features = [f for f in records if f['counted_feature']]
        milestone = [f for f in features if f['milestone_scope'] == 'linux-cli']
        for name, rows in [('record_counts', records), ('counts', features), ('milestone_counts', milestone)]:
            self.assertEqual(self.result[name], dict(collections.Counter(f['disposition'] for f in rows)))
        self.assertEqual(self.result['coverage_records'], len(records))
        self.assertEqual(self.result['semantic_features'], len(features))
        self.assertEqual(self.result['milestone_features'], len(milestone))
        self.assertGreater(len(records), len(features))
        self.assertGreater(len(features), len(milestone))
        self.assertFalse(any(not f['counted_feature'] for f in records if f['disposition'] in ('mapping-pending', 'validator-declared')))

    def test_lsp_fields_have_separate_artifact_evidence(self):
        artifacts = self.result['native_artifacts']['copilot']
        fields = [f for f in artifacts if f['feature'].startswith('artifact:lsp:/')]
        self.assertEqual(len(fields), 14)
        for field in fields:
            self.assertIn(field['scope'], ('project', 'user'))
            self.assertEqual(field['native_status'], 'bounded-fixture-execution')
            if field['feature'].endswith('/requestTimeoutMs'):
                self.assertTrue(any('timeout' in path for path in field['evidence']))
            self.assertTrue(field['evidence'])
        self.assertFalse(any('lspServers' in f['path'] for f in self.registry['copilot'].values() if isinstance(f, dict)))

    def test_generated_file_is_current(self):
        self.assertEqual(self.result, json.loads((ROOT / '.agents/features/coverage.json').read_text()))


if __name__ == '__main__':
    unittest.main()
