import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Button, 
  Space, 
  Typography, 
  Input, 
  Alert, 
  Spin, 
  Row, 
  Col,
  Select,
  Modal,
  Form,
  Steps,
  Divider,
  Badge,
  Tooltip,
  DatePicker,
  notification
} from 'antd';
import {
  SafetyOutlined,
  ReloadOutlined,
  FileProtectOutlined,
  SendOutlined,
  SyncOutlined,
  EyeOutlined,
  InfoCircleOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  CopyOutlined,
  CloudOutlined
} from '@ant-design/icons';
// Note: You may need to install dayjs with: npm install dayjs
// import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { Step } = Steps;
const { TextArea } = Input;

// Types matching the Go backend
interface Certificate {
  name: string;
  conf: string;
  issuer: string;
  tags: string[];
  cloudflare: {
    path: string;
    zone: string;
  };
  secret: {
    platform: string;
    env: string;
    path: string;
    keys: Record<string, string>;
  };
}

interface CertificateOperation {
  success: boolean;
  message: string;
  output: string;
}

const Certificates: React.FC = () => {
  const [certificates, setCertificates] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedCertificate, setSelectedCertificate] = useState<string>('');
  const [jiraTicket, setJiraTicket] = useState<string>('');
  const [operationResult, setOperationResult] = useState<CertificateOperation | null>(null);
  const [currentStep, setCurrentStep] = useState(0);
  
  // Modal states
  const [showRenewModal, setShowRenewModal] = useState(false);
  const [showRefreshModal, setShowRefreshModal] = useState(false);
  const [showDescribeModal, setShowDescribeModal] = useState(false);
  const [showNotificationModal, setShowNotificationModal] = useState(false);
  
  // Form states
  const [renewForm] = Form.useForm();
  const [refreshForm] = Form.useForm();
  const [describeForm] = Form.useForm();
  const [notificationForm] = Form.useForm();
  
  // Operation states
  const [renewLoading, setRenewLoading] = useState(false);
  const [refreshLoading, setRefreshLoading] = useState(false);
  const [describeLoading, setDescribeLoading] = useState(false);
  const [gandiTokenStatus, setGandiTokenStatus] = useState<CertificateOperation | null>(null);

  useEffect(() => {
    loadCertificates();
    checkGandiToken();
  }, []);

  const loadCertificates = async () => {
    setLoading(true);
    setError(null);
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const certList = await window.go.main.App.ListCertificates();
        setCertificates(certList);
      }
    } catch (error) {
      console.error('Failed to load certificates:', error);
      setError(`Failed to load certificates: ${error instanceof Error ? error.message : String(error)}`);
    } finally {
      setLoading(false);
    }
  };

  const checkGandiToken = async () => {
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const result = await window.go.main.App.CheckGandiToken();
        setGandiTokenStatus(result);
      }
    } catch (error) {
      console.error('Failed to check Gandi token:', error);
      setGandiTokenStatus({
        success: false,
        message: 'Failed to check Gandi token',
        output: error instanceof Error ? error.message : String(error)
      });
    }
  };

  const handleRenewCertificate = async (values: any) => {
    setRenewLoading(true);
    try {
      const result = await window.go.main.App.RenewCertificate(values.certificate, values.jiraTicket);
      setOperationResult(result);
      if (result.success) {
        notification.success({
          message: 'Certificate Renewal Initiated',
          description: result.message,
        });
        setShowRenewModal(false);
        renewForm.resetFields();
      } else {
        notification.error({
          message: 'Certificate Renewal Failed',
          description: result.message,
        });
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      notification.error({
        message: 'Error',
        description: `Failed to renew certificate: ${errorMessage}`,
      });
    } finally {
      setRenewLoading(false);
    }
  };

  const handleRefreshSecret = async (values: any) => {
    setRefreshLoading(true);
    try {
      const result = await window.go.main.App.RefreshCertificateSecret(values.certificate, values.jiraTicket);
      setOperationResult(result);
      if (result.success) {
        notification.success({
          message: 'Secret Refreshed',
          description: result.message,
        });
        setShowRefreshModal(false);
        refreshForm.resetFields();
      } else {
        notification.error({
          message: 'Secret Refresh Failed',
          description: result.message,
        });
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      notification.error({
        message: 'Error',
        description: `Failed to refresh secret: ${errorMessage}`,
      });
    } finally {
      setRefreshLoading(false);
    }
  };

  const handleDescribeSecret = async (values: any) => {
    setDescribeLoading(true);
    try {
      const result = await window.go.main.App.DescribeCertificateSecret(
        values.certificate, 
        values.version || 0, 
        values.diffVersion || 0
      );
      setOperationResult(result);
      if (result.success) {
        notification.success({
          message: 'Certificate Details Retrieved',
          description: result.message,
        });
      } else {
        notification.error({
          message: 'Failed to Get Details',
          description: result.message,
        });
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      notification.error({
        message: 'Error',
        description: `Failed to describe certificate: ${errorMessage}`,
      });
    } finally {
      setDescribeLoading(false);
    }
  };

  const handleSendNotification = async (values: any) => {
    try {
      const result = await window.go.main.App.SendCertificateNotification(
        values.certificate,
        values.operationDate.format ? values.operationDate.format('YYYY-MM-DD') : values.operationDate,
        values.operation
      );
      setOperationResult(result);
      notification.success({
        message: 'Notification Template Generated',
        description: result.message,
      });
      setShowNotificationModal(false);
      notificationForm.resetFields();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      notification.error({
        message: 'Error',
        description: `Failed to send notification: ${errorMessage}`,
      });
    }
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      notification.success({
        message: 'Copied to clipboard',
        description: 'Email content has been copied to your clipboard',
      });
    } catch (error) {
      notification.error({
        message: 'Copy failed',
        description: 'Failed to copy to clipboard',
      });
    }
  };

  const steps = [
    {
      title: 'Pre-checks',
      description: 'Verify token and timing',
      icon: <InfoCircleOutlined />,
    },
    {
      title: 'Generate Notification',
      description: 'Prepare email for Technical Services',
      icon: <SendOutlined />,
    },
    {
      title: 'Renew Certificate',
      description: 'Request renewal from provider',
      icon: <SyncOutlined />,
    },
    {
      title: 'Review & Merge PR',
      description: 'Approve DNS validation changes',
      icon: <CheckCircleOutlined />,
    },
    {
      title: 'Wait for Domain Validation',
      description: 'Wait for Gandi email confirmation',
      icon: <WarningOutlined />,
    },
    {
      title: 'Refresh Secret',
      description: 'Update vault with new certificate',
      icon: <FileProtectOutlined />,
    },
    {
      title: 'Review & Merge PR',
      description: 'Approve secret version bump',
      icon: <CheckCircleOutlined />,
    },
    {
      title: 'Terraform Apply',
      description: 'Deploy to workspaces',
      icon: <CloudOutlined />,
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Row justify="space-between" style={{ marginBottom: '24px' }}>
        <Col>
          <Title level={2}>
            <FileProtectOutlined style={{ marginRight: '8px' }} />
            SSL Certificate Management
          </Title>
          <Text type="secondary">Manage SSL certificate renewals using yak commands</Text>
        </Col>
        <Col>
          <Space>
            <Button 
              icon={<ReloadOutlined />}
              onClick={loadCertificates}
              loading={loading}
            >
              Refresh
            </Button>
          </Space>
        </Col>
      </Row>

      {error && (
        <Alert
          message={error}
          type="error"
          showIcon
          closable
          onClose={() => setError(null)}
          style={{ marginBottom: '16px' }}
        />
      )}

      {/* Gandi Token Status */}
      <Card title="Configuration Status" style={{ marginBottom: '16px' }}>
        <Row gutter={16}>
          <Col span={12}>
            <Space>
              <Text strong>GANDI_TOKEN:</Text>
              {gandiTokenStatus ? (
                <Badge 
                  status={gandiTokenStatus.success ? 'success' : 'error'} 
                  text={gandiTokenStatus.success ? 'Valid' : 'Invalid'} 
                />
              ) : (
                <Spin size="small" />
              )}
              <Button size="small" onClick={checkGandiToken}>
                Recheck
              </Button>
            </Space>
            {gandiTokenStatus && !gandiTokenStatus.success && (
              <Alert
                message="Gandi Token Issue"
                description={gandiTokenStatus.output}
                type="warning"
                size="small"
                style={{ marginTop: '8px' }}
              />
            )}
          </Col>
        </Row>
      </Card>

      {/* Process Steps */}
      <Card title="Certificate Renewal Process" style={{ marginBottom: '16px' }}>
        <Steps current={currentStep} size="small">
          {steps.map((step, index) => (
            <Step 
              key={index}
              title={step.title}
              description={step.description}
              icon={step.icon}
            />
          ))}
        </Steps>
        
        <Divider />
        
        <Alert
          message="Important Warning"
          description="Once certificate validation is done, you have ONLY 48 HOURS to deploy the new certificate because the old one is revoked by Gandi after 48 hours."
          type="warning"
          showIcon
          icon={<WarningOutlined />}
          style={{ marginTop: '16px' }}
        />
      </Card>

      {/* Action Buttons */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Space direction="vertical" style={{ width: '100%' }}>
              <SendOutlined style={{ fontSize: '24px', color: '#1890ff' }} />
              <Text strong>Generate Notification</Text>
              <Text type="secondary">Prepare email template for Technical Services</Text>
              <Button 
                type="primary" 
                block 
                onClick={() => setShowNotificationModal(true)}
              >
                Generate Notification
              </Button>
            </Space>
          </Card>
        </Col>
        
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Space direction="vertical" style={{ width: '100%' }}>
              <SyncOutlined style={{ fontSize: '24px', color: '#52c41a' }} />
              <Text strong>Renew Certificate</Text>
              <Text type="secondary">Request renewal from provider</Text>
              <Button 
                type="primary" 
                block 
                onClick={() => setShowRenewModal(true)}
                disabled={!gandiTokenStatus?.success}
              >
                Start Renewal
              </Button>
            </Space>
          </Card>
        </Col>
        
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Space direction="vertical" style={{ width: '100%' }}>
              <FileProtectOutlined style={{ fontSize: '24px', color: '#faad14' }} />
              <Text strong>Refresh Secret</Text>
              <Text type="secondary">Update vault with new certificate</Text>
              <Button 
                type="primary" 
                block 
                onClick={() => setShowRefreshModal(true)}
                disabled={!gandiTokenStatus?.success}
              >
                Refresh Secret
              </Button>
            </Space>
          </Card>
        </Col>
        
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Space direction="vertical" style={{ width: '100%' }}>
              <EyeOutlined style={{ fontSize: '24px', color: '#722ed1' }} />
              <Text strong>Describe Secret</Text>
              <Text type="secondary">View certificate details</Text>
              <Button 
                type="primary" 
                block 
                onClick={() => setShowDescribeModal(true)}
                disabled={!gandiTokenStatus?.success}
              >
                View Details
              </Button>
            </Space>
          </Card>
        </Col>
      </Row>

      {/* Domain Validation Waiting Step */}
      <Card 
        title={
          <Space>
            <WarningOutlined style={{ color: '#faad14' }} />
            <Text strong>Critical: Wait for Domain Validation</Text>
          </Space>
        } 
        style={{ marginTop: '16px' }}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Alert
            message="Important Waiting Period"
            description="After merging the DNS validation PR, you must wait for Gandi to complete domain validation. This can take several hours."
            type="warning"
            showIcon
          />
          
          <Alert
            message="⚠️ 48-Hour Critical Window Starts Now!"
            description="Once you receive the domain validation email from Gandi, you have ONLY 48 hours to complete the refresh secret step and deploy the certificate before the old one is revoked."
            type="error"
            showIcon
          />
        </Space>
      </Card>

      {/* Final Step - Terraform Apply */}
      <Card 
        title={
          <Space>
            <CloudOutlined style={{ color: '#52c41a' }} />
            <Text strong>Final Step: Terraform Apply</Text>
          </Space>
        } 
        style={{ marginTop: '16px' }}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Alert
            message="After the PR is merged"
            description="Once the secret version has been bumped, you need to trigger a run in TFE for all workspaces that use the certificate."
            type="info"
            showIcon
          />
          
          <div>
            <Text strong>Find workspace usage with this command in terraform-infra repository:</Text>
            <pre style={{ 
              background: '#f5f5f5', 
              padding: '12px', 
              borderRadius: '4px',
              fontSize: '12px',
              marginTop: '8px',
              border: '1px solid #d9d9d9'
            }}>
{`# replace with your secret vault path
grep -r 'common/wildcard-certs/doctolib-net-wildcard' **/*.tf`}
            </pre>
          </div>
          
          <Text type="secondary">
            Then manually trigger terraform apply in TFE for each workspace that uses the certificate.
          </Text>
        </Space>
      </Card>

      {/* Operation Result */}
      {operationResult && (
        <Card title="Operation Result" style={{ marginTop: '16px' }}>
          <Alert
            message={operationResult.message}
            type={operationResult.success ? 'success' : 'error'}
            showIcon
            style={{ marginBottom: '16px' }}
          />
          {operationResult.output && (
            <div>
              <Row justify="space-between" align="middle" style={{ marginBottom: '8px' }}>
                <Col>
                  <Text strong>Output:</Text>
                </Col>
                <Col>
                  <Button 
                    size="small" 
                    icon={<CopyOutlined />}
                    onClick={() => copyToClipboard(operationResult.output)}
                  >
                    Copy
                  </Button>
                </Col>
              </Row>
              <pre style={{ 
                background: '#f5f5f5', 
                padding: '12px', 
                borderRadius: '4px',
                fontSize: '12px',
                whiteSpace: 'pre-wrap',
                border: '1px solid #d9d9d9'
              }}>
                {operationResult.output}
              </pre>
            </div>
          )}
        </Card>
      )}

      {/* Generate Notification Modal */}
      <Modal
        title="Generate Notification for Technical Services"
        open={showNotificationModal}
        onCancel={() => setShowNotificationModal(false)}
        footer={null}
        width={600}
      >
        <Form
          form={notificationForm}
          layout="vertical"
          onFinish={handleSendNotification}
        >
          <Form.Item
            name="certificate"
            label="Certificate"
            rules={[{ required: true, message: 'Please select a certificate' }]}
          >
            <Select placeholder="Select certificate">
              {certificates.map(cert => (
                <Option key={cert} value={cert}>{cert}</Option>
              ))}
            </Select>
          </Form.Item>
          
          <Form.Item
            name="operation"
            label="Operation Type"
            rules={[{ required: true, message: 'Please select operation type' }]}
          >
            <Select placeholder="Select operation type">
              <Option value="renewal">Certificate Renewal</Option>
              <Option value="refresh">Secret Refresh</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            name="operationDate"
            label="Scheduled Date"
            rules={[{ required: true, message: 'Please select operation date' }]}
          >
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                Generate Template
              </Button>
              <Button onClick={() => setShowNotificationModal(false)}>
                Cancel
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Renew Certificate Modal */}
      <Modal
        title="Renew Certificate"
        open={showRenewModal}
        onCancel={() => setShowRenewModal(false)}
        footer={null}
        width={600}
      >
        <Alert
          message="Before Starting"
          description="Make sure it's not too early to renew (can be renewed up to 30 days before expiration) and that you have notified the Technical Services team."
          type="info"
          showIcon
          style={{ marginBottom: '16px' }}
        />
        
        <Card size="small" style={{ marginBottom: '16px', backgroundColor: '#fafafa' }}>
          <Text strong>This command will:</Text>
          <ul style={{ marginTop: '8px', marginBottom: '0', paddingLeft: '20px' }}>
            <li style={{ marginBottom: '4px' }}>• Generate the CSR from the private key in the secret configured</li>
            <li style={{ marginBottom: '4px' }}>• Ask Gandi for a renewal</li>
            <li style={{ marginBottom: '4px' }}>• Attach tags of the old certificate to the new certificate</li>
            <li style={{ marginBottom: '4px' }}>• Build the pull request to add Cloudflare DNS records (for Gandi domain control validation)</li>
          </ul>
        </Card>
        
        <Alert
          message="Next Step"
          description="After running this command, you will need to review the pull request and merge it to master before proceeding."
          type="warning"
          showIcon
          style={{ marginBottom: '16px' }}
        />
        
        <Form
          form={renewForm}
          layout="vertical"
          onFinish={handleRenewCertificate}
        >
          <Form.Item
            name="certificate"
            label="Certificate Name"
            rules={[{ required: true, message: 'Please select a certificate' }]}
          >
            <Select placeholder="Select certificate">
              {certificates.map(cert => (
                <Option key={cert} value={cert}>{cert}</Option>
              ))}
            </Select>
          </Form.Item>
          
          <Form.Item
            name="jiraTicket"
            label="JIRA Ticket"
            rules={[{ required: true, message: 'Please enter JIRA ticket' }]}
          >
            <Input placeholder="e.g., JIRA-1234" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={renewLoading}>
                Start Renewal
              </Button>
              <Button onClick={() => setShowRenewModal(false)}>
                Cancel
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Refresh Secret Modal */}
      <Modal
        title="Refresh Certificate Secret"
        open={showRefreshModal}
        onCancel={() => setShowRefreshModal(false)}
        footer={null}
        width={600}
      >
        <Alert
          message="Prerequisites"
          description="Make sure the certificate renewal has been completed and domain validation is finished before running this command."
          type="warning"
          showIcon
          style={{ marginBottom: '16px' }}
        />
        
        <Card size="small" style={{ marginBottom: '16px', backgroundColor: '#fafafa' }}>
          <Text strong>This command will:</Text>
          <ul style={{ marginTop: '8px', marginBottom: '0', paddingLeft: '20px' }}>
            <li style={{ marginBottom: '4px' }}>• Get the new valid certificate (if any, otherwise ask for the domain control validation again)</li>
            <li style={{ marginBottom: '4px' }}>• Sanity checks on the certificate (is expiration good, do we have the good private key in the secret...)</li>
            <li style={{ marginBottom: '4px' }}>• Combine Intermediate certificate with the new certificate</li>
            <li style={{ marginBottom: '4px' }}>• Update the secret</li>
            <li style={{ marginBottom: '4px' }}>• Build the pull request(s) to bump secrets</li>
          </ul>
        </Card>
        
        <Alert
          message="Next Step"
          description="After running this command, you will need to review the pull request and merge it to master before proceeding."
          type="warning"
          showIcon
          style={{ marginBottom: '16px' }}
        />
        
        <Form
          form={refreshForm}
          layout="vertical"
          onFinish={handleRefreshSecret}
        >
          <Form.Item
            name="certificate"
            label="Certificate Name"
            rules={[{ required: true, message: 'Please select a certificate' }]}
          >
            <Select placeholder="Select certificate">
              {certificates.map(cert => (
                <Option key={cert} value={cert}>{cert}</Option>
              ))}
            </Select>
          </Form.Item>
          
          <Form.Item
            name="jiraTicket"
            label="JIRA Ticket"
            rules={[{ required: true, message: 'Please enter JIRA ticket' }]}
          >
            <Input placeholder="e.g., JIRA-1234" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={refreshLoading}>
                Refresh Secret
              </Button>
              <Button onClick={() => setShowRefreshModal(false)}>
                Cancel
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Describe Secret Modal */}
      <Modal
        title="Describe Certificate Secret"
        open={showDescribeModal}
        onCancel={() => setShowDescribeModal(false)}
        footer={null}
        width={600}
      >
        <Form
          form={describeForm}
          layout="vertical"
          onFinish={handleDescribeSecret}
        >
          <Form.Item
            name="certificate"
            label="Certificate Name"
            rules={[{ required: true, message: 'Please select a certificate' }]}
          >
            <Select placeholder="Select certificate">
              {certificates.map(cert => (
                <Option key={cert} value={cert}>{cert}</Option>
              ))}
            </Select>
          </Form.Item>
          
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="version"
                label="Version (optional)"
              >
                <Input type="number" placeholder="e.g., 2" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="diffVersion"
                label="Diff Version (optional)"
              >
                <Input type="number" placeholder="e.g., 1" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={describeLoading}>
                Get Details
              </Button>
              <Button onClick={() => setShowDescribeModal(false)}>
                Cancel
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Certificates;