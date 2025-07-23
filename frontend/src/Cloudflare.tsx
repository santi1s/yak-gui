import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Tabs, 
  Typography, 
  Form, 
  Input, 
  Button, 
  Space, 
  Table, 
  message, 
  Modal, 
  Select,
  Switch,
  Tooltip,
  Tag,
  Descriptions,
  Divider
} from 'antd';
import { 
  CloudOutlined, 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined,
  ReloadOutlined,
  SettingOutlined,
  InfoCircleOutlined,
  CopyOutlined,
  SearchOutlined
} from '@ant-design/icons';

const { Title, Text } = Typography;
const { Option } = Select;
const { TabPane } = Tabs;

interface CloudflareConfig {
  token: string;
  accountName: string;
}

interface Zone {
  id: string;
  name: string;
  status: string;
  paused: boolean;
  type: string;
  development_mode?: boolean;
  name_servers?: string[];
}

interface DNSRecord {
  id: string;
  type: string;
  name: string;
  content: string;
  ttl: number;
  proxied?: boolean;
  zone_id: string;
  zone_name: string;
  created_on?: string;
  modified_on?: string;
}

interface LoadBalancer {
  id: string;
  name: string;
  description?: string;
  enabled: boolean;
  ttl?: number;
  proxied: boolean;
  region_pools?: any;
  pop_pools?: any;
  country_pools?: any;
  default_pools: string[];
  fallback_pool?: string;
  session_affinity: string;
  steering_policy: string;
  created_on?: string;
  modified_on?: string;
  session_affinity_attributes?: {
    samesite?: string;
    secure?: string;
    zero_downtime_failover?: string;
  };
  random_steering?: {
    default_weight?: number;
  };
  adaptive_routing?: {
    failover_across_pools?: boolean;
  };
  location_strategy?: {
    prefer_ecs?: string;
    mode?: string;
  };
}

interface Pool {
  id: string;
  name: string;
  description?: string;
  enabled: boolean;
  healthy?: boolean;
  minimum_origins: number;
  origins: Origin[];
  monitor?: string;
  notification_email?: string;
  created_on?: string;
  modified_on?: string;
  origin_steering?: {
    policy?: string;
  };
  check_regions?: any;
}

interface Origin {
  name: string;
  address: string;
  enabled: boolean;
  weight?: number;
  header?: any;
}

interface WAFRule {
  id: string;
  description: string;
  expression: string;
  action: string;
  enabled: boolean;
  ref?: string;
  version?: string;
}

const Cloudflare: React.FC = () => {
  const [form] = Form.useForm();
  const [config, setConfig] = useState<CloudflareConfig | null>(null);
  const [zones, setZones] = useState<Zone[]>([]);
  const [dnsRecords, setDnsRecords] = useState<DNSRecord[]>([]);
  const [loadBalancers, setLoadBalancers] = useState<LoadBalancer[]>([]);
  const [pools, setPools] = useState<Pool[]>([]);
  const [wafRules, setWafRules] = useState<WAFRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [configModalVisible, setConfigModalVisible] = useState(false);
  const [selectedZone, setSelectedZone] = useState<string>('');
  const [activeTab, setActiveTab] = useState('dns');
  const [dnsFilter, setDnsFilter] = useState<{name: string, type: string}>({name: '', type: ''});
  const [filteredDnsRecords, setFilteredDnsRecords] = useState<DNSRecord[]>([]);
  const [lbDetailVisible, setLbDetailVisible] = useState(false);
  const [selectedLb, setSelectedLb] = useState<LoadBalancer | null>(null);
  const [poolFilter, setPoolFilter] = useState<{name: string, id: string}>({name: '', id: ''});
  const [filteredPools, setFilteredPools] = useState<Pool[]>([]);
  const [poolDetailVisible, setPoolDetailVisible] = useState(false);
  const [selectedPool, setSelectedPool] = useState<Pool | null>(null);

  useEffect(() => {
    loadConfig();
  }, []);

  useEffect(() => {
    if (config && config.token && config.accountName) {
      loadZones();
    }
  }, [config]);

  useEffect(() => {
    if (selectedZone && config) {
      loadDataForTab(activeTab);
    }
  }, [selectedZone, activeTab, config]);

  useEffect(() => {
    // Filter DNS records based on name and type
    let filtered = dnsRecords;
    
    if (dnsFilter.name) {
      filtered = filtered.filter(record => 
        record.name.toLowerCase().includes(dnsFilter.name.toLowerCase())
      );
    }
    
    if (dnsFilter.type) {
      filtered = filtered.filter(record => 
        record.type.toLowerCase() === dnsFilter.type.toLowerCase()
      );
    }
    
    setFilteredDnsRecords(filtered);
  }, [dnsRecords, dnsFilter]);

  useEffect(() => {
    // Filter pools based on name and ID
    let filtered = pools;
    
    if (poolFilter.name) {
      filtered = filtered.filter(pool => 
        pool.name.toLowerCase().includes(poolFilter.name.toLowerCase())
      );
    }
    
    if (poolFilter.id) {
      filtered = filtered.filter(pool => 
        pool.id.toLowerCase().includes(poolFilter.id.toLowerCase())
      );
    }
    
    setFilteredPools(filtered);
  }, [pools, poolFilter]);

  const loadConfig = async () => {
    try {
      const response = await window.go.main.App.GetCloudflareConfig();
      // Set default account name if not already set
      const configWithDefaults = {
        ...response,
        accountName: response.accountName || 'Doctolib'
      };
      setConfig(configWithDefaults);
    } catch (error) {
      console.error('Failed to load Cloudflare config:', error);
    }
  };

  const saveConfig = async (values: CloudflareConfig) => {
    try {
      setLoading(true);
      await window.go.main.App.SetCloudflareConfig(values);
      setConfig(values);
      setConfigModalVisible(false);
      message.success('Configuration saved successfully');
      loadZones();
    } catch (error) {
      console.error('Failed to save config:', error);
      message.error('Failed to save configuration');
    } finally {
      setLoading(false);
    }
  };

  const loadZones = async () => {
    if (!config?.token || !config?.accountName) return;
    
    try {
      setLoading(true);
      const zoneList = await window.go.main.App.GetCloudflareZones(config);
      setZones(zoneList);
      if (zoneList.length > 0 && !selectedZone) {
        setSelectedZone(zoneList[0].name);
      }
    } catch (error) {
      console.error('Failed to load zones:', error);
      message.error('Failed to load zones');
    } finally {
      setLoading(false);
    }
  };

  const loadDataForTab = async (tab: string) => {
    if (!config || !selectedZone) return;

    try {
      setLoading(true);
      
      switch (tab) {
        case 'dns':
          const records = await window.go.main.App.GetCloudflareDNSRecords(config, selectedZone);
          setDnsRecords(records);
          break;
        case 'lb':
          const lbs = await window.go.main.App.GetCloudflareLoadBalancers(config, selectedZone);
          setLoadBalancers(lbs);
          break;
        case 'pool':
          const poolList = await window.go.main.App.GetCloudflarePools(config);
          setPools(poolList);
          break;
        case 'waf':
          const rules = await window.go.main.App.GetCloudflareWAFRules(config, selectedZone);
          setWafRules(rules);
          break;
      }
    } catch (error) {
      console.error(`Failed to load ${tab} data:`, error);
      message.error(`Failed to load ${tab} data`);
    } finally {
      setLoading(false);
    }
  };

  const createDNSRecord = async (record: Partial<DNSRecord>) => {
    if (!config || !selectedZone) return;
    
    try {
      setLoading(true);
      await window.go.main.App.CreateCloudflareDNSRecord(config, selectedZone, record);
      message.success('DNS record created successfully');
      loadDataForTab('dns');
    } catch (error) {
      console.error('Failed to create DNS record:', error);
      message.error('Failed to create DNS record');
    } finally {
      setLoading(false);
    }
  };

  const deleteDNSRecord = async (recordId: string) => {
    if (!config || !selectedZone) return;
    
    try {
      setLoading(true);
      await window.go.main.App.DeleteCloudflareDNSRecord(config, selectedZone, recordId);
      message.success('DNS record deleted successfully');
      loadDataForTab('dns');
    } catch (error) {
      console.error('Failed to delete DNS record:', error);
      message.error('Failed to delete DNS record');
    } finally {
      setLoading(false);
    }
  };

  const navigateToPoolById = (poolId: string) => {
    // Close the load balancer detail modal
    setLbDetailVisible(false);
    // Switch to pools tab
    setActiveTab('pool');
    // Set the pool ID filter
    setPoolFilter({name: '', id: poolId});
  };


  const dnsColumns = [
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      width: 80,
      render: (type: string) => <Tag color="blue">{type}</Tag>
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: 'Content',
      dataIndex: 'content',
      key: 'content',
      ellipsis: true,
    },
    {
      title: 'TTL',
      dataIndex: 'ttl',
      key: 'ttl',
      width: 80,
      render: (ttl: number) => ttl === 1 ? 'Auto' : ttl.toString()
    },
    {
      title: 'Proxied',
      dataIndex: 'proxied',
      key: 'proxied',
      width: 80,
      render: (proxied: boolean) => 
        <Tag color={proxied ? 'orange' : 'default'}>
          {proxied ? 'Yes' : 'DNS only'}
        </Tag>
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      render: (_: any, record: DNSRecord) => (
        <Space>
          <Button 
            type="link" 
            icon={<EditOutlined />} 
            size="small"
            onClick={() => {/* TODO: Implement edit */}}
          />
          <Button 
            type="link" 
            icon={<DeleteOutlined />} 
            danger 
            size="small"
            onClick={() => deleteDNSRecord(record.id)}
          />
        </Space>
      ),
    },
  ];

  const loadBalancerColumns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: 'Status',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 100,
      render: (enabled: boolean) => 
        <Tag color={enabled ? 'green' : 'red'}>
          {enabled ? 'Enabled' : 'Disabled'}
        </Tag>
    },
    {
      title: 'Proxied',
      dataIndex: 'proxied',
      key: 'proxied',
      width: 80,
      render: (proxied: boolean) => 
        <Tag color={proxied ? 'orange' : 'default'}>
          {proxied ? 'Yes' : 'No'}
        </Tag>
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      render: (_: any, record: LoadBalancer) => (
        <Button 
          type="link" 
          icon={<InfoCircleOutlined />} 
          size="small"
          onClick={() => {
            setSelectedLb(record);
            setLbDetailVisible(true);
          }}
        >
          Details
        </Button>
      ),
    },
  ];

  const poolColumns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      sorter: (a: Pool, b: Pool) => a.name.localeCompare(b.name),
      defaultSortOrder: 'ascend' as const,
    },
    {
      title: 'Pool ID',
      dataIndex: 'id',
      key: 'id',
      width: 200,
      ellipsis: true,
      render: (id: string) => (
        <Text code copyable style={{ fontSize: '11px' }}>
          {id}
        </Text>
      ),
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (desc: string) => desc || 'No description',
    },
    {
      title: 'Status',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 100,
      render: (enabled: boolean) => 
        <Tag color={enabled ? 'green' : 'red'}>
          {enabled ? 'Enabled' : 'Disabled'}
        </Tag>
    },
    {
      title: 'Health',
      dataIndex: 'healthy',
      key: 'healthy',
      width: 100,
      render: (healthy: boolean | undefined) => {
        if (healthy === undefined) return <Tag color="default">Unknown</Tag>;
        return (
          <Tag color={healthy ? 'green' : 'red'}>
            {healthy ? 'Healthy' : 'Unhealthy'}
          </Tag>
        );
      }
    },
    {
      title: 'Origins',
      dataIndex: 'origins',
      key: 'origins',
      width: 80,
      render: (origins: Origin[]) => origins?.length || 0
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      render: (_: any, record: Pool) => (
        <Button 
          type="link" 
          icon={<InfoCircleOutlined />} 
          size="small"
          onClick={() => {
            setSelectedPool(record);
            setPoolDetailVisible(true);
          }}
        >
          Details
        </Button>
      ),
    },
  ];

  const wafColumns = [
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: 'Expression',
      dataIndex: 'expression',
      key: 'expression',
      ellipsis: true,
    },
    {
      title: 'Action',
      dataIndex: 'action',
      key: 'action',
      render: (action: string) => 
        <Tag color={action === 'block' ? 'red' : action === 'allow' ? 'green' : 'blue'}>
          {action}
        </Tag>
    },
    {
      title: 'Status',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean) => 
        <Tag color={enabled ? 'green' : 'red'}>
          {enabled ? 'Enabled' : 'Disabled'}
        </Tag>
    },
  ];

  if (!config?.token || !config?.accountName) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '60px 20px' }}>
          <CloudOutlined style={{ fontSize: '64px', color: '#1890ff', marginBottom: '16px' }} />
          <Title level={3}>Cloudflare Configuration Required</Title>
          <Text type="secondary">
            Please configure your Cloudflare API token and account name to get started.
          </Text>
          <br />
          <Button 
            type="primary" 
            icon={<SettingOutlined />}
            style={{ marginTop: '16px' }}
            onClick={() => setConfigModalVisible(true)}
          >
            Configure Cloudflare
          </Button>
        </div>
        
        <Modal
          title="Cloudflare Configuration"
          open={configModalVisible}
          onCancel={() => setConfigModalVisible(false)}
          footer={null}
        >
          <Form
            form={form}
            layout="vertical"
            onFinish={saveConfig}
            initialValues={{ accountName: 'Doctolib' }}
          >
            <Form.Item
              label={
                <Space>
                  API Token
                  <Tooltip title="Cloudflare API token with appropriate permissions">
                    <InfoCircleOutlined />
                  </Tooltip>
                </Space>
              }
              name="token"
              rules={[{ required: true, message: 'Please input your Cloudflare API token!' }]}
            >
              <Input.Password placeholder="Your Cloudflare API token" />
            </Form.Item>
            
            <Form.Item
              label="Account Name"
              name="accountName"
              rules={[{ required: true, message: 'Please input your account name!' }]}
            >
              <Input placeholder="Your Cloudflare account name" />
            </Form.Item>
            
            <Form.Item>
              <Space>
                <Button type="primary" htmlType="submit" loading={loading}>
                  Save Configuration
                </Button>
                <Button onClick={() => setConfigModalVisible(false)}>
                  Cancel
                </Button>
              </Space>
            </Form.Item>
          </Form>
        </Modal>
      </Card>
    );
  }

  return (
    <Card>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
        <Title level={4} style={{ margin: 0 }}>
          <CloudOutlined style={{ marginRight: '8px', color: '#1890ff' }} />
          Cloudflare Management
        </Title>
        <Space>
          <Select
            value={selectedZone}
            onChange={setSelectedZone}
            style={{ minWidth: '200px' }}
            loading={loading}
            placeholder="Select Zone"
          >
            {zones.map(zone => (
              <Option key={zone.id} value={zone.name}>
                {zone.name}
              </Option>
            ))}
          </Select>
          <Button 
            icon={<ReloadOutlined />} 
            onClick={() => loadDataForTab(activeTab)}
            loading={loading}
          >
            Refresh
          </Button>
          <Button 
            icon={<SettingOutlined />} 
            onClick={() => setConfigModalVisible(true)}
          >
            Settings
          </Button>
        </Space>
      </div>

      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="DNS Records" key="dns">
          <div style={{ marginBottom: '16px', display: 'flex', gap: '12px', alignItems: 'center' }}>
            <Button 
              type="primary" 
              icon={<PlusOutlined />}
              onClick={() => {/* TODO: Implement create DNS record modal */}}
            >
              Create DNS Record
            </Button>
            <Input
              placeholder="Filter by name..."
              value={dnsFilter.name}
              onChange={(e) => setDnsFilter(prev => ({ ...prev, name: e.target.value }))}
              style={{ width: '200px' }}
              allowClear
            />
            <Select
              placeholder="Filter by type..."
              value={dnsFilter.type || undefined}
              onChange={(value) => setDnsFilter(prev => ({ ...prev, type: value || '' }))}
              style={{ width: '150px' }}
              allowClear
            >
              <Option value="A">A</Option>
              <Option value="AAAA">AAAA</Option>
              <Option value="CNAME">CNAME</Option>
              <Option value="MX">MX</Option>
              <Option value="TXT">TXT</Option>
              <Option value="NS">NS</Option>
              <Option value="SRV">SRV</Option>
              <Option value="PTR">PTR</Option>
              <Option value="CAA">CAA</Option>
            </Select>
            <Text type="secondary" style={{ marginLeft: 'auto' }}>
              {filteredDnsRecords.length} of {dnsRecords.length} records
            </Text>
          </div>
          <Table
            columns={dnsColumns}
            dataSource={filteredDnsRecords}
            loading={loading}
            rowKey="id"
            size="small"
          />
        </TabPane>

        <TabPane tab="Load Balancers" key="lb">
          <Table
            columns={loadBalancerColumns}
            dataSource={loadBalancers}
            loading={loading}
            rowKey="id"
            size="small"
          />
        </TabPane>

        <TabPane tab="Pools" key="pool">
          <div style={{ marginBottom: '16px', display: 'flex', gap: '12px', alignItems: 'center' }}>
            <Input
              placeholder="Filter by pool name..."
              value={poolFilter.name}
              onChange={(e) => setPoolFilter(prev => ({ ...prev, name: e.target.value }))}
              style={{ width: '200px' }}
              allowClear
            />
            <Input
              placeholder="Filter by pool ID..."
              value={poolFilter.id}
              onChange={(e) => setPoolFilter(prev => ({ ...prev, id: e.target.value }))}
              style={{ width: '250px' }}
              allowClear
            />
            <Text type="secondary" style={{ marginLeft: 'auto' }}>
              {filteredPools.length} of {pools.length} pools
            </Text>
          </div>
          <Table
            columns={poolColumns}
            dataSource={filteredPools}
            loading={loading}
            rowKey="id"
            size="small"
          />
        </TabPane>

        <TabPane tab="WAF Rules" key="waf">
          <Table
            columns={wafColumns}
            dataSource={wafRules}
            loading={loading}
            rowKey="id"
            size="small"
          />
        </TabPane>
      </Tabs>

      <Modal
        title="Cloudflare Configuration"
        open={configModalVisible}
        onCancel={() => setConfigModalVisible(false)}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{...config, accountName: config?.accountName || 'Doctolib'}}
          onFinish={saveConfig}
        >
          <Form.Item
            label={
              <Space>
                API Token
                <Tooltip title="Cloudflare API token with appropriate permissions">
                  <InfoCircleOutlined />
                </Tooltip>
              </Space>
            }
            name="token"
            rules={[{ required: true, message: 'Please input your Cloudflare API token!' }]}
          >
            <Input.Password placeholder="Your Cloudflare API token" />
          </Form.Item>
          
          <Form.Item
            label="Account Name"
            name="accountName"
            rules={[{ required: true, message: 'Please input your account name!' }]}
          >
            <Input placeholder="Your Cloudflare account name" />
          </Form.Item>
          
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                Save Configuration
              </Button>
              <Button onClick={() => setConfigModalVisible(false)}>
                Cancel
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Load Balancer Detail Modal */}
      <Modal
        title={
          <Space>
            <InfoCircleOutlined />
            <span>Load Balancer Details: {selectedLb?.name}</span>
          </Space>
        }
        open={lbDetailVisible}
        onCancel={() => setLbDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setLbDetailVisible(false)}>
            Close
          </Button>
        ]}
        width={800}
      >
        {selectedLb && (
          <div style={{ maxHeight: '600px', overflow: 'auto' }}>
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="ID" span={2}>
                <Text code copyable>{selectedLb.id}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Name" span={2}>
                <Text strong>{selectedLb.name}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Description" span={2}>
                {selectedLb.description || 'No description'}
              </Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color={selectedLb.enabled ? 'green' : 'red'}>
                  {selectedLb.enabled ? 'Enabled' : 'Disabled'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Proxied">
                <Tag color={selectedLb.proxied ? 'orange' : 'default'}>
                  {selectedLb.proxied ? 'Yes' : 'No'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Session Affinity">
                <Tag>{selectedLb.session_affinity}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Steering Policy">
                <Tag color="blue">{selectedLb.steering_policy}</Tag>
              </Descriptions.Item>
              {selectedLb.created_on && (
                <Descriptions.Item label="Created">
                  {new Date(selectedLb.created_on).toLocaleString()}
                </Descriptions.Item>
              )}
              {selectedLb.modified_on && (
                <Descriptions.Item label="Modified">
                  {new Date(selectedLb.modified_on).toLocaleString()}
                </Descriptions.Item>
              )}
            </Descriptions>

            <Divider>Pool Configuration</Divider>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="Default Pools">
                {selectedLb.default_pools && selectedLb.default_pools.length > 0 ? (
                  <div>
                    {selectedLb.default_pools.map((poolId, index) => {
                      const pool = pools.find(p => p.id === poolId);
                      return (
                        <div key={index} style={{ marginBottom: '4px' }}>
                          <Button
                            type="link"
                            size="small"
                            style={{ padding: '0 4px', fontSize: '11px', fontFamily: 'monospace' }}
                            onClick={() => navigateToPoolById(poolId)}
                            title={pool ? `Click to view pool: ${pool.name}` : 'Click to search for this pool ID'}
                          >
                            {poolId}
                          </Button>
                          <Button
                            type="text"
                            size="small"
                            icon={<CopyOutlined />}
                            style={{ marginLeft: '4px', padding: '0 4px' }}
                            onClick={() => navigator.clipboard.writeText(poolId)}
                            title="Copy pool ID"
                          />
                          {pool && (
                            <Text type="secondary" style={{ marginLeft: '8px', fontSize: '10px' }}>
                              ({pool.name})
                            </Text>
                          )}
                        </div>
                      );
                    })}
                  </div>
                ) : (
                  'No default pools'
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Fallback Pool">
                {selectedLb.fallback_pool ? (
                  <div>
                    <Button
                      type="link"
                      size="small"
                      style={{ padding: '0 4px', fontSize: '11px', fontFamily: 'monospace' }}
                      onClick={() => navigateToPoolById(selectedLb.fallback_pool!)}
                      title={(() => {
                        const pool = pools.find(p => p.id === selectedLb.fallback_pool);
                        return pool ? `Click to view pool: ${pool.name}` : 'Click to search for this pool ID';
                      })()}
                    >
                      {selectedLb.fallback_pool}
                    </Button>
                    <Button
                      type="text"
                      size="small"
                      icon={<CopyOutlined />}
                      style={{ marginLeft: '4px', padding: '0 4px' }}
                      onClick={() => navigator.clipboard.writeText(selectedLb.fallback_pool!)}
                      title="Copy pool ID"
                    />
                    {(() => {
                      const pool = pools.find(p => p.id === selectedLb.fallback_pool);
                      return pool && (
                        <Text type="secondary" style={{ marginLeft: '8px', fontSize: '10px' }}>
                          ({pool.name})
                        </Text>
                      );
                    })()}
                  </div>
                ) : (
                  'No fallback pool'
                )}
              </Descriptions.Item>
            </Descriptions>

            {selectedLb.session_affinity_attributes && Object.keys(selectedLb.session_affinity_attributes).length > 0 && (
              <>
                <Divider>Session Affinity Attributes</Divider>
                <Descriptions column={2} bordered size="small">
                  <Descriptions.Item label="SameSite">
                    {selectedLb.session_affinity_attributes?.samesite || 'Not set'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Secure">
                    {selectedLb.session_affinity_attributes?.secure || 'Not set'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Zero Downtime Failover" span={2}>
                    {selectedLb.session_affinity_attributes?.zero_downtime_failover || 'Not set'}
                  </Descriptions.Item>
                </Descriptions>
              </>
            )}

            {selectedLb.random_steering && Object.keys(selectedLb.random_steering).length > 0 && (
              <>
                <Divider>Random Steering</Divider>
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="Default Weight">
                    {selectedLb.random_steering?.default_weight || 'Not set'}
                  </Descriptions.Item>
                </Descriptions>
              </>
            )}

            {selectedLb.adaptive_routing && Object.keys(selectedLb.adaptive_routing).length > 0 && (
              <>
                <Divider>Adaptive Routing</Divider>
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="Failover Across Pools">
                    <Tag color={selectedLb.adaptive_routing?.failover_across_pools ? 'green' : 'red'}>
                      {selectedLb.adaptive_routing?.failover_across_pools ? 'Enabled' : 'Disabled'}
                    </Tag>
                  </Descriptions.Item>
                </Descriptions>
              </>
            )}

            {selectedLb.location_strategy && Object.keys(selectedLb.location_strategy).length > 0 && (
              <>
                <Divider>Location Strategy</Divider>
                <Descriptions column={2} bordered size="small">
                  <Descriptions.Item label="Prefer ECS">
                    {selectedLb.location_strategy?.prefer_ecs || 'Not set'}
                  </Descriptions.Item>
                  <Descriptions.Item label="Mode">
                    {selectedLb.location_strategy?.mode || 'Not set'}
                  </Descriptions.Item>
                </Descriptions>
              </>
            )}
          </div>
        )}
      </Modal>

      {/* Pool Detail Modal */}
      <Modal
        title={
          <Space>
            <InfoCircleOutlined />
            <span>Pool Details: {selectedPool?.name}</span>
          </Space>
        }
        open={poolDetailVisible}
        onCancel={() => setPoolDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setPoolDetailVisible(false)}>
            Close
          </Button>
        ]}
        width={900}
      >
        {selectedPool && (
          <div style={{ maxHeight: '600px', overflow: 'auto' }}>
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="ID" span={2}>
                <Text code copyable>{selectedPool.id}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Name" span={2}>
                <Text strong>{selectedPool.name}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Description" span={2}>
                {selectedPool.description || 'No description'}
              </Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color={selectedPool.enabled ? 'green' : 'red'}>
                  {selectedPool.enabled ? 'Enabled' : 'Disabled'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Health">
                {selectedPool.healthy === undefined ? (
                  <Tag color="default">Unknown</Tag>
                ) : (
                  <Tag color={selectedPool.healthy ? 'green' : 'red'}>
                    {selectedPool.healthy ? 'Healthy' : 'Unhealthy'}
                  </Tag>
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Minimum Origins">
                <Text code>{selectedPool.minimum_origins}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Total Origins">
                <Text code>{selectedPool.origins?.length || 0}</Text>
              </Descriptions.Item>
              {selectedPool.created_on && (
                <Descriptions.Item label="Created">
                  {new Date(selectedPool.created_on).toLocaleString()}
                </Descriptions.Item>
              )}
              {selectedPool.modified_on && (
                <Descriptions.Item label="Modified">
                  {new Date(selectedPool.modified_on).toLocaleString()}
                </Descriptions.Item>
              )}
            </Descriptions>

            <Divider>Origins</Divider>
            {selectedPool.origins && selectedPool.origins.length > 0 ? (
              <Table
                columns={[
                  {
                    title: 'Name',
                    dataIndex: 'name',
                    key: 'name',
                    ellipsis: true,
                  },
                  {
                    title: 'Address',
                    dataIndex: 'address',
                    key: 'address',
                    ellipsis: true,
                  },
                  {
                    title: 'Status',
                    dataIndex: 'enabled',
                    key: 'enabled',
                    width: 100,
                    render: (enabled: boolean) => 
                      <Tag color={enabled ? 'green' : 'red'}>
                        {enabled ? 'Enabled' : 'Disabled'}
                      </Tag>
                  },
                  {
                    title: 'Weight',
                    dataIndex: 'weight',
                    key: 'weight',
                    width: 80,
                    render: (weight: number) => weight || 1,
                  },
                  {
                    title: 'Actions',
                    key: 'actions',
                    width: 100,
                    render: (_: any, record: Origin) => (
                      <Button 
                        type="link" 
                        icon={<SearchOutlined />} 
                        size="small"
                        onClick={async () => {
                          try {
                            const result = await window.go.main.App.ResolveHostname(record.address);
                            Modal.info({
                              title: `DNS Resolution: ${record.address}`,
                              content: (
                                <pre style={{ 
                                  fontFamily: 'monospace', 
                                  fontSize: '12px',
                                  background: '#f5f5f5',
                                  padding: '12px',
                                  borderRadius: '4px',
                                  margin: '12px 0'
                                }}>
                                  {result}
                                </pre>
                              ),
                              width: 700,
                            });
                          } catch (error: any) {
                            Modal.error({
                              title: `DNS Resolution Failed`,
                              content: `Could not resolve ${record.address}: ${error.message || error}`,
                            });
                          }
                        }}
                        title={`Resolve DNS for ${record.address}`}
                      >
                        resolve
                      </Button>
                    ),
                  },
                ]}
                dataSource={selectedPool.origins}
                pagination={false}
                size="small"
                rowKey="name"
              />
            ) : (
              <Text type="secondary">No origins configured</Text>
            )}

            {selectedPool.origin_steering && Object.keys(selectedPool.origin_steering).length > 0 && (
              <>
                <Divider>Origin Steering</Divider>
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="Policy">
                    <Tag color="blue">{selectedPool.origin_steering?.policy || 'Not set'}</Tag>
                  </Descriptions.Item>
                </Descriptions>
              </>
            )}

            {selectedPool.monitor && (
              <>
                <Divider>Monitoring</Divider>
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="Monitor ID">
                    <Text code copyable style={{ fontSize: '11px' }}>{selectedPool.monitor}</Text>
                  </Descriptions.Item>
                  {selectedPool.notification_email && (
                    <Descriptions.Item label="Notification Email">
                      <Text code>{selectedPool.notification_email}</Text>
                    </Descriptions.Item>
                  )}
                </Descriptions>
              </>
            )}

            {selectedPool.check_regions && (
              <>
                <Divider>Check Regions</Divider>
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="Regions">
                    {Array.isArray(selectedPool.check_regions) ? 
                      selectedPool.check_regions.join(', ') : 
                      JSON.stringify(selectedPool.check_regions)
                    }
                  </Descriptions.Item>
                </Descriptions>
              </>
            )}
          </div>
        )}
      </Modal>

    </Card>
  );
};

export default Cloudflare;