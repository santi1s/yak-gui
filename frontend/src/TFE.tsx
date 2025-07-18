import React, { useState, useEffect } from 'react';
import {
  Layout,
  Card,
  Button,
  Input,
  Select,
  Space,
  Typography,
  Alert,
  Spin,
  Table,
  Tag,
  Row,
  Col,
  Tabs,
  Form,
  Switch,
  Tooltip,
  Progress,
  Modal,
  List,
  Divider,
} from 'antd';
import {
  CloudServerOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  StopOutlined,
  ReloadOutlined,
  SettingOutlined,
  FileTextOutlined,
  DatabaseOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  LockOutlined,
  UnlockOutlined,
  DeleteOutlined,
  EditOutlined,
  SearchOutlined,
  UserOutlined,
  CopyOutlined,
} from '@ant-design/icons';

const { Title, Text } = Typography;
const { Option } = Select;

// TFE interfaces based on the yak tfe command analysis
interface TFEWorkspace {
  id: string;
  name: string;
  description?: string;
  environment?: string;
  terraformVersion?: string;
  status?: 'active' | 'locked' | 'disabled';
  lastRun?: string;
  owner?: string;
  tags?: string[];
  organization?: string;
  createdAt?: string;
  updatedAt?: string;
  autoApply?: boolean;
  terraformWorking?: boolean;
  vcsRepo?: {
    identifier: string;
    branch: string;
    ingressSubmodules: boolean;
  };
  variables?: Record<string, string>;
  executionMode?: string;
  workspaceType?: string;
  vcsConnection?: {
    repository: string;
    branch: string;
    workingDirectory: string;
    webhookURL: string;
  };
}

interface TFERun {
  id: string;
  workspaceId: string;
  workspaceName: string;
  status: 'pending' | 'planning' | 'planned' | 'applying' | 'applied' | 'discarded' | 'errored' | 'canceled';
  createdAt: string;
  message?: string;
  source?: string;
  terraformVersion?: string;
  hasChanges?: boolean;
  isDestroy?: boolean;
  isConfirmable?: boolean;
  actions?: {
    isConfirmable: boolean;
    isCancelable: boolean;
    isDiscardable: boolean;
  };
  createdBy?: string;
  url?: string;
}

interface TFEConfig {
  endpoint: string;
  organization: string;
  token?: string;
}

// Helper function to extract environment from workspace name (same logic as backend)
const extractEnvironmentFromName = (name: string): string => {
  // Check for regional patterns first (more specific)
  if (name.includes('dev-aws-fr-par-1')) return 'dev-aws-fr-par-1';
  if (name.includes('dev-aws-de-fra-1')) return 'dev-aws-de-fra-1';
  if (name.includes('dev-aws-global')) return 'dev-aws-global';
  if (name.includes('staging-aws-fr-par-1')) return 'staging-aws-fr-par-1';
  if (name.includes('staging-aws-de-fra-1')) return 'staging-aws-de-fra-1';
  if (name.includes('staging-aws-global')) return 'staging-aws-global';
  if (name.includes('prod-aws-fr-par-1') || name.includes('prd-aws-fr-par-1')) return 'prod-aws-fr-par-1';
  if (name.includes('prod-aws-de-fra-1') || name.includes('prd-aws-de-fra-1')) return 'prod-aws-de-fra-1';
  if (name.includes('prod-aws-global') || name.includes('prd-aws-global')) return 'prod-aws-global';
  if (name.includes('preprod-aws-fr-par-1')) return 'preprod-aws-fr-par-1';
  if (name.includes('preprod-aws-de-fra-1')) return 'preprod-aws-de-fra-1';
  if (name.includes('preprod-aws-global')) return 'preprod-aws-global';
  
  // Check for general environment patterns
  if (name.includes('preprod')) return 'preprod';
  if (name.includes('shared')) return 'shared';
  if (name.includes('prd') || name.includes('prod')) return 'production';
  if (name.includes('staging')) return 'staging';
  if (name.includes('dev')) return 'development';
  if (name.includes('test')) return 'testing';
  
  return 'unknown';
};

const TFE: React.FC = () => {
  const [config, setConfig] = useState<TFEConfig>({
    endpoint: 'app.terraform.io',
    organization: 'doctolib',
  });
  const [workspaces, setWorkspaces] = useState<TFEWorkspace[]>([]);
  const [filteredWorkspaces, setFilteredWorkspaces] = useState<TFEWorkspace[]>([]);
  const [runs, setRuns] = useState<TFERun[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedWorkspace, setSelectedWorkspace] = useState<string | null>(null);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [activeTab, setActiveTab] = useState<'workspaces' | 'runs' | 'versions'>('workspaces');
  const [showConfigModal, setShowConfigModal] = useState(false);
  const [nameFilter, setNameFilter] = useState('');
  const [environmentFilter, setEnvironmentFilter] = useState('');
  const [tagFilter, setTagFilter] = useState('');
  const [negateTagFilter, setNegateTagFilter] = useState(false);
  const [pageSize, setPageSize] = useState(10);
  const [configForm] = Form.useForm();
  const [terraformVersions, setTerraformVersions] = useState<any[]>([]);
  const [versionSummary, setVersionSummary] = useState<any>(null);
  const [runStatusFilter, setRunStatusFilter] = useState('');
  const [selectedRuns, setSelectedRuns] = useState<string[]>([]);
  const [runDateFilter, setRunDateFilter] = useState<string>('');
  const [workspaceVariables, setWorkspaceVariables] = useState<any[]>([]);
  const [selectedWorkspaceForVars, setSelectedWorkspaceForVars] = useState<string>('');
  const [variableSets, setVariableSets] = useState<any[]>([]);
  const [selectedVariableSet, setSelectedVariableSet] = useState<string>('');
  const [variableViewMode, setVariableViewMode] = useState<'workspace' | 'variable-set'>('workspace');
  const [includeVariableSets, setIncludeVariableSets] = useState(false);
  const [variableSetDetails, setVariableSetDetails] = useState<any>(null);
  const [selectedWorkspaceDetails, setSelectedWorkspaceDetails] = useState<any>(null);
  const [selectedWorkspaceForDetails, setSelectedWorkspaceForDetails] = useState<string>('');
  const [runLogs, setRunLogs] = useState<string>('');
  const [selectedRunForLogs, setSelectedRunForLogs] = useState<string>('');
  const [runLogsVisible, setRunLogsVisible] = useState(false);
  const [runLogsLoading, setRunLogsLoading] = useState(false);

  // Load workspaces
  const loadWorkspaces = async () => {
    if (!config?.endpoint || !config?.organization) {
      console.warn('TFE config not ready, skipping workspace load');
      return;
    }
    
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.GetTFEWorkspaces) {
        const workspaces = await window.go.main.App.GetTFEWorkspaces(config);
        setWorkspaces(workspaces || []);
        setFilteredWorkspaces(workspaces || []);
      } else {
        throw new Error('TFE backend not available');
      }
    } catch (error) {
      console.error('Failed to load workspaces:', error);
      setError(`Failed to load workspaces: ${error}`);
      setWorkspaces([]);
      setFilteredWorkspaces([]);
    } finally {
      setLoading(false);
    }
  };

  // Load enhanced workspace details (for inline display, not modal)
  const loadWorkspaceDetails = async (workspaceName: string) => {
    if (!workspaceName) {
      setError('No workspace name provided');
      return;
    }
    
    setLoading(true);
    setError(null);
    setSelectedWorkspaceDetails(null);
    
    try {
      if (window.go?.main?.App?.GetTFEWorkspaceDetails) {
        const details = await window.go.main.App.GetTFEWorkspaceDetails(config, workspaceName);
        console.log('Workspace details loaded:', details);
        setSelectedWorkspaceDetails(details || null);
        // Don't show modal - we display details inline now
        // setWorkspaceDetailsVisible(true);
      } else {
        setError('Backend function not available');
      }
    } catch (error) {
      console.error('Failed to load workspace details:', error);
      setError(`Failed to load workspace details: ${error}`);
      setSelectedWorkspaceDetails(null);
    } finally {
      setLoading(false);
    }
  };

  // Handle workspace selection for details view
  const handleWorkspaceSelection = async (workspaceName: string) => {
    setSelectedWorkspaceForDetails(workspaceName);
    if (workspaceName) {
      await loadWorkspaceDetails(workspaceName);
    } else {
      setSelectedWorkspaceDetails(null);
    }
  };

  // Load run logs
  const loadRunLogs = async (runId: string) => {
    if (!runId) {
      setError('No run ID provided');
      return;
    }
    
    setRunLogsLoading(true);
    setRunLogs('');
    setSelectedRunForLogs(runId);
    
    try {
      if (window.go?.main?.App?.GetTFERunLogs) {
        const logs = await window.go.main.App.GetTFERunLogs(config, runId);
        setRunLogs(logs || 'No logs available');
        setRunLogsVisible(true);
      } else {
        setError('TFE run logs backend not available');
      }
    } catch (error) {
      console.error('Failed to load run logs:', error);
      setError(`Failed to load run logs: ${error}`);
      setRunLogs('Failed to load logs');
    } finally {
      setRunLogsLoading(false);
    }
  };

  // Single workspace actions
  const lockWorkspace = async (workspaceName: string) => {
    if (!workspaceName) return;
    
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.LockTFEWorkspace) {
        await window.go.main.App.LockTFEWorkspace(config, [workspaceName], true);
        // Reload workspace details to reflect the change
        await handleWorkspaceSelection(workspaceName);
      } else {
        throw new Error('TFE workspace lock backend not available');
      }
    } catch (error) {
      setError(`Failed to lock workspace: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const unlockWorkspace = async (workspaceName: string) => {
    if (!workspaceName) return;
    
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.UnlockTFEWorkspace) {
        await window.go.main.App.UnlockTFEWorkspace(config, [workspaceName], false);
        // Reload workspace details to reflect the change
        await handleWorkspaceSelection(workspaceName);
      } else {
        throw new Error('TFE workspace unlock backend not available');
      }
    } catch (error) {
      setError(`Failed to unlock workspace: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const setWorkspaceVersion = async (workspaceName: string, version: string) => {
    if (!workspaceName || !version) return;
    
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.SetTFEWorkspaceVersion) {
        await window.go.main.App.SetTFEWorkspaceVersion(config, [workspaceName], version);
        // Reload workspace details to reflect the change
        await handleWorkspaceSelection(workspaceName);
      } else {
        throw new Error('TFE workspace version setting backend not available');
      }
    } catch (error) {
      setError(`Failed to set workspace version: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  // Load workspaces with real TFE tag filtering
  const loadWorkspacesByTag = async (tag: string, not: boolean = false) => {
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.GetTFEWorkspacesByTag) {
        const workspaces = await window.go.main.App.GetTFEWorkspacesByTag(config, tag, not);
        setWorkspaces(workspaces);
        setFilteredWorkspaces(workspaces);
      } else {
        throw new Error('TFE tag filtering backend not available');
      }
    } catch (error) {
      setError(`Failed to load workspaces by tag: ${error}`);
      setWorkspaces([]);
      setFilteredWorkspaces([]);
    } finally {
      setLoading(false);
    }
  };

  // Load runs - requires workspace selection
  const loadRuns = async (workspaceId?: string) => {
    const targetWorkspace = workspaceId || selectedWorkspace;
    if (!targetWorkspace) {
      setError('Please select a workspace to load runs');
      return;
    }
    
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.GetTFERuns) {
        const runs = await window.go.main.App.GetTFERuns(config, targetWorkspace);
        setRuns(runs);
      } else {
        throw new Error('TFE runs backend not available');
      }
    } catch (error) {
      setError(`Failed to load runs: ${error}`);
      setRuns([]);
    } finally {
      setLoading(false);
    }
  };

  // Enhanced run management functions
  const discardOldRuns = async (ageHours: number, discardPending: boolean = false) => {
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.DiscardTFERuns) {
        await window.go.main.App.DiscardTFERuns(config, ageHours, discardPending, false, false);
        await loadRuns(); // Refresh runs after discarding
      } else {
        throw new Error('TFE discard runs backend not available');
      }
    } catch (error) {
      setError(`Failed to discard old runs: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const bulkDiscardRuns = async () => {
    if (selectedRuns.length === 0) return;
    
    setLoading(true);
    setError(null);
    try {
      // Note: This would need a specific backend function for bulk run operations
      // For now, we'll use the existing discard function
      if (window.go?.main?.App?.DiscardTFERuns) {
        await window.go.main.App.DiscardTFERuns(config, 0, false, false, false);
        await loadRuns(); // Refresh runs after operation
        setSelectedRuns([]); // Clear selection
      } else {
        throw new Error('TFE bulk discard runs backend not available');
      }
    } catch (error) {
      setError(`Failed to bulk discard runs: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  // Filter runs based on status and date
  const getFilteredRuns = () => {
    let filtered = [...runs];

    // Filter by status
    if (runStatusFilter) {
      filtered = filtered.filter(run => run.status === runStatusFilter);
    }

    // Filter by date (last 24 hours, 7 days, etc.)
    if (runDateFilter) {
      const now = new Date();
      const filterDate = new Date();
      
      switch (runDateFilter) {
        case '24h':
          filterDate.setHours(now.getHours() - 24);
          break;
        case '7d':
          filterDate.setDate(now.getDate() - 7);
          break;
        case '30d':
          filterDate.setDate(now.getDate() - 30);
          break;
        default:
          return filtered;
      }
      
      filtered = filtered.filter(run => new Date(run.created_at) >= filterDate);
    }

    return filtered;
  };

  const getUniqueRunStatuses = () => {
    const statuses = runs
      .map(run => run.status)
      .filter((status, index, self) => status && self.indexOf(status) === index)
      .sort();
    return statuses;
  };



  // Advanced tag-based workspace loading
  const loadWorkspacesByTagAdvanced = async () => {
    if (!tagFilter) {
      await loadWorkspaces();
      return;
    }
    
    if (!config?.endpoint || !config?.organization) {
      setError('TFE configuration not ready');
      return;
    }
    
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.GetTFEWorkspacesByTag) {
        const workspaces = await window.go.main.App.GetTFEWorkspacesByTag(config, tagFilter, negateTagFilter);
        setWorkspaces(workspaces || []);
        setFilteredWorkspaces(workspaces || []);
      } else {
        throw new Error('TFE tag filtering backend not available');
      }
    } catch (error) {
      console.error('Tag filtering error:', error);
      setError(`Failed to load workspaces by tag: ${error}`);
      setWorkspaces([]);
      setFilteredWorkspaces([]);
    } finally {
      setLoading(false);
    }
  };

  // Load TFE configuration
  const loadTFEConfig = async () => {
    try {
      if (window.go?.main?.App?.GetTFEConfig) {
        const tfeConfig = await window.go.main.App.GetTFEConfig();
        if (tfeConfig) {
          setConfig(tfeConfig);
          if (configForm) {
            configForm.setFieldsValue(tfeConfig);
          }
        }
      }
    } catch (error) {
      console.warn('Failed to load TFE config:', error);
    }
  };

  // Filter workspace names based on multiple criteria
  const filterWorkspaces = () => {
    let filtered = [...workspaces];

    // Filter by name
    if (nameFilter) {
      filtered = filtered.filter(workspace =>
        workspace.name.toLowerCase().includes(nameFilter.toLowerCase())
      );
    }

    // Filter by environment (extracted from workspace name)
    if (environmentFilter) {
      filtered = filtered.filter(workspace =>
        extractEnvironmentFromName(workspace.name) === environmentFilter
      );
    }

    // Note: We can't filter by owner, status, or tags since we only have workspace names
    // These filters will be applied server-side via tag filtering

    setFilteredWorkspaces(filtered);
  };

  // Get unique values for filter dropdowns (from workspace names)
  const getUniqueEnvironments = () => {
    const environments = workspaces
      .map(workspace => extractEnvironmentFromName(workspace.name))
      .filter((env, index, self) => env && env !== 'unknown' && self.indexOf(env) === index)
      .sort();
    return environments;
  };

  // Helper function to extract environment from workspace name (moved from backend)
  const extractEnvironmentFromName = (name: string) => {
    // Check for regional patterns first (more specific)
    if (name.includes('dev-aws-fr-par-1')) return 'dev-aws-fr-par-1';
    if (name.includes('dev-aws-de-fra-1')) return 'dev-aws-de-fra-1';
    if (name.includes('dev-aws-global')) return 'dev-aws-global';
    if (name.includes('staging-aws-fr-par-1')) return 'staging-aws-fr-par-1';
    if (name.includes('staging-aws-de-fra-1')) return 'staging-aws-de-fra-1';
    if (name.includes('staging-aws-global')) return 'staging-aws-global';
    if (name.includes('prod-aws-fr-par-1') || name.includes('prd-aws-fr-par-1')) return 'prod-aws-fr-par-1';
    if (name.includes('prod-aws-de-fra-1') || name.includes('prd-aws-de-fra-1')) return 'prod-aws-de-fra-1';
    if (name.includes('prod-aws-global') || name.includes('prd-aws-global')) return 'prod-aws-global';
    if (name.includes('preprod-aws-fr-par-1')) return 'preprod-aws-fr-par-1';
    if (name.includes('preprod-aws-de-fra-1')) return 'preprod-aws-de-fra-1';
    if (name.includes('preprod-aws-global')) return 'preprod-aws-global';
    
    // Check for general environment patterns
    if (name.includes('preprod')) return 'preprod';
    if (name.includes('shared')) return 'shared';
    if (name.includes('prd') || name.includes('prod')) return 'production';
    if (name.includes('staging')) return 'staging';
    if (name.includes('dev')) return 'development';
    if (name.includes('test')) return 'testing';
    
    return 'unknown';
  };

  // Handle configuration save
  const handleConfigSave = async (values: TFEConfig) => {
    try {
      if (window.go?.main?.App?.SetTFEConfig) {
        await window.go.main.App.SetTFEConfig(values);
      }
      setConfig(values);
      setShowConfigModal(false);
      // Reload workspaces with new config
      await loadWorkspaces();
    } catch (error) {
      setError(`Failed to save configuration: ${error}`);
    }
  };

  // Version management functions
  const loadTerraformVersions = async () => {
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.GetTFEVersions) {
        const versions = await window.go.main.App.GetTFEVersions(config);
        setTerraformVersions(versions);
        
        // Calculate version summary
        const summary = calculateVersionSummary(versions);
        setVersionSummary(summary);
      } else {
        throw new Error('TFE versions backend not available');
      }
    } catch (error) {
      setError(`Failed to load Terraform versions: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const checkDeprecatedVersions = async () => {
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.CheckTFEDeprecatedVersions) {
        const result = await window.go.main.App.CheckTFEDeprecatedVersions(config, '', '', false);
        // Handle the result - could show a modal or update the UI
        console.log('Deprecated versions check result:', result);
      } else {
        throw new Error('TFE deprecated versions check backend not available');
      }
    } catch (error) {
      setError(`Failed to check deprecated versions: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const calculateVersionSummary = (versions: any[]) => {
    const workspaceVersions = workspaces.map(ws => ws.terraform_version).filter(v => v);
    const totalWorkspaces = workspaceVersions.length;
    
    const versionCounts = workspaceVersions.reduce((acc, version) => {
      acc[version] = (acc[version] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);
    
    const deprecatedVersions = versions.filter(v => v.status === 'deprecated' || v.isDeprecated);
    const deprecatedCount = workspaceVersions.filter(v => 
      deprecatedVersions.some(dv => dv.version === v)
    ).length;
    
    const topVersions = Object.entries(versionCounts)
      .sort(([,a], [,b]) => b - a)
      .slice(0, 5)
      .map(([version, count]) => ({
        version,
        count,
        isDeprecated: deprecatedVersions.some(dv => dv.version === version)
      }));
    
    return {
      totalWorkspaces,
      deprecatedCount,
      currentCount: totalWorkspaces - deprecatedCount,
      topVersions
    };
  };

  // Variable management functions
  const loadWorkspaceVariables = async (workspaceId: string) => {
    if (!workspaceId) {
      setError('No workspace selected');
      return;
    }
    
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.GetTFEWorkspaceVariables) {
        const variables = await window.go.main.App.GetTFEWorkspaceVariables(config, workspaceId, includeVariableSets);
        setWorkspaceVariables(variables || []);
      } else {
        throw new Error('TFE workspace variables backend not available');
      }
    } catch (error) {
      console.error('Failed to load workspace variables:', error);
      setError(`Failed to load workspace variables: ${error}`);
      setWorkspaceVariables([]);
    } finally {
      setLoading(false);
    }
  };

  const loadVariableSetVariables = async (variableSetName: string) => {
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.GetTFEVariableSetVariables) {
        const variables = await window.go.main.App.GetTFEVariableSetVariables(config, variableSetName);
        setWorkspaceVariables(variables);
      } else {
        throw new Error('TFE variable set variables backend not available');
      }
    } catch (error) {
      setError(`Failed to load variable set variables: ${error}`);
      setWorkspaceVariables([]);
    } finally {
      setLoading(false);
    }
  };

  const loadVariableSetDetails = async (variableSetName: string) => {
    if (!variableSetName) {
      setError('No variable set selected');
      return;
    }
    
    setLoading(true);
    setError(null);
    try {
      if (window.go?.main?.App?.GetTFEVariableSetDetails) {
        const details = await window.go.main.App.GetTFEVariableSetDetails(config, variableSetName);
        setVariableSetDetails(details || null);
        setWorkspaceVariables((details && details.variables) || []);
      } else {
        throw new Error('TFE variable set details backend not available');
      }
    } catch (error) {
      console.error('Failed to load variable set details:', error);
      setError(`Failed to load variable set details: ${error}`);
      setVariableSetDetails(null);
      setWorkspaceVariables([]);
    } finally {
      setLoading(false);
    }
  };

  const loadVariableSets = async () => {
    if (!config?.endpoint || !config?.organization) {
      console.warn('TFE config not ready, skipping variable sets load');
      setVariableSets([]);
      return;
    }
    
    try {
      if (window.go?.main?.App?.GetTFEVariableSets) {
        console.log('Loading variable sets with config:', config);
        const sets = await window.go.main.App.GetTFEVariableSets(config);
        console.log('Variable sets loaded:', sets);
        setVariableSets(sets || []);
      } else {
        console.warn('TFE variable sets backend not available');
        setVariableSets([]);
      }
    } catch (error) {
      console.error('Failed to load variable sets:', error);
      setVariableSets([]);
      // Don't set global error for variable sets failure
    }
  };

  useEffect(() => {
    const initializeTFE = async () => {
      await loadTFEConfig();
      // Auto-load workspaces and variable sets after config is loaded
      setTimeout(() => {
        loadWorkspaces();
        loadVariableSets();
      }, 200);
    };
    
    initializeTFE();
  }, []);

  useEffect(() => {
    if (autoRefresh) {
      const interval = setInterval(() => {
        loadWorkspaces();
        if (selectedWorkspace) {
          loadRuns();
        }
      }, 30000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, selectedWorkspace]);

  // Filter workspaces when filters change (only name and environment available for workspace names)
  useEffect(() => {
    filterWorkspaces();
  }, [nameFilter, environmentFilter, workspaces]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'green';
      case 'locked': return 'orange';
      case 'disabled': return 'red';
      case 'applied': return 'green';
      case 'planned': return 'blue';
      case 'planning': return 'processing';
      case 'applying': return 'processing';
      case 'errored': return 'red';
      case 'canceled': return 'default';
      default: return 'default';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active': return <CheckCircleOutlined />;
      case 'locked': return <LockOutlined />;
      case 'disabled': return <StopOutlined />;
      case 'applied': return <CheckCircleOutlined />;
      case 'planned': return <InfoCircleOutlined />;
      case 'errored': return <ExclamationCircleOutlined />;
      default: return <InfoCircleOutlined />;
    }
  };


  const runColumns = [
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (text: string) => (
        <Tag color={getStatusColor(text)} icon={getStatusIcon(text)}>
          {text}
        </Tag>
      ),
    },
    {
      title: 'Message',
      dataIndex: 'message',
      key: 'message',
      ellipsis: true,
    },
    {
      title: 'Source',
      dataIndex: 'source',
      key: 'source',
      render: (text: string) => <Tag>{text}</Tag>,
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text: string) => {
        if (!text) return 'N/A';
        try {
          const date = new Date(text);
          if (isNaN(date.getTime())) return 'Invalid Date';
          return date.toLocaleString();
        } catch (error) {
          return 'Invalid Date';
        }
      },
    },
    {
      title: 'Changes',
      dataIndex: 'hasChanges',
      key: 'hasChanges',
      render: (hasChanges: boolean) => (
        <Tag color={hasChanges ? 'orange' : 'green'}>
          {hasChanges ? 'Has Changes' : 'No Changes'}
        </Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (text: any, record: TFERun) => (
        <Space size="middle">
          <Tooltip title="View run logs">
            <Button
              type="text"
              icon={<FileTextOutlined />}
              onClick={() => loadRunLogs(record.id)}
              loading={runLogsLoading && selectedRunForLogs === record.id}
              size="small"
            />
          </Tooltip>
          {record.url && (
            <Tooltip title="Open in TFE">
              <Button
                type="text"
                icon={<ExclamationCircleOutlined />}
                onClick={() => window.open(record.url, '_blank')}
                size="small"
              />
            </Tooltip>
          )}
        </Space>
      ),
    },
  ];

  const versionColumns = [
    {
      title: 'Version',
      dataIndex: 'version',
      key: 'version',
      render: (text: string) => <Text code>{text}</Text>,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (text: string) => (
        <Tag color={text === 'deprecated' ? 'red' : text === 'supported' ? 'green' : 'blue'}>
          {text}
        </Tag>
      ),
    },
    {
      title: 'Usage',
      dataIndex: 'usage',
      key: 'usage',
      render: (usage: number) => (
        <Text>{usage} workspace{usage !== 1 ? 's' : ''}</Text>
      ),
    },
    {
      title: 'Default',
      dataIndex: 'isDefault',
      key: 'isDefault',
      render: (isDefault: boolean) => (
        isDefault ? <Tag color="blue">Default</Tag> : null
      ),
    },
    {
      title: 'Beta',
      dataIndex: 'beta',
      key: 'beta',
      render: (beta: boolean) => (
        beta ? <Tag color="orange">Beta</Tag> : null
      ),
    },
  ];

  const tabItems = [
    {
      key: 'workspaces',
      label: (
        <span>
          <DatabaseOutlined />
          Workspaces
        </span>
      ),
      children: (
        <div>
          <Row justify="space-between" style={{ marginBottom: '16px' }}>
            <Col>
              <Space>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={loadWorkspaces}
                  loading={loading}
                >
                  Refresh
                </Button>
                <Switch
                  checked={autoRefresh}
                  onChange={setAutoRefresh}
                  checkedChildren="Auto"
                  unCheckedChildren="Manual"
                />
              </Space>
            </Col>
            <Col>
              <Text type="secondary">
                {filteredWorkspaces.length} of {workspaces.length} workspaces
              </Text>
            </Col>
          </Row>
          
          {/* Filter Controls */}
          <Card size="small" style={{ marginBottom: '16px' }}>
            <Row gutter={[16, 16]}>
              <Col xs={24} sm={8} md={6}>
                <Input
                  placeholder="Filter by workspace name..."
                  value={nameFilter}
                  onChange={(e) => setNameFilter(e.target.value)}
                  prefix={<SearchOutlined />}
                  allowClear
                  autoCapitalize="none"
                  autoCorrect="off"
                  spellCheck={false}
                />
              </Col>
              <Col xs={24} sm={8} md={6}>
                <Select
                  placeholder="Filter by environment..."
                  value={environmentFilter}
                  onChange={setEnvironmentFilter}
                  allowClear
                  style={{ width: '100%' }}
                >
                  <Option value="">All Environments</Option>
                  {getUniqueEnvironments().map(env => (
                    <Option key={env} value={env}>
                      {env}
                    </Option>
                  ))}
                </Select>
              </Col>
              <Col xs={24} sm={8} md={12}>
                <Space>
                  <Input
                    placeholder="Filter by tag..."
                    value={tagFilter}
                    onChange={(e) => setTagFilter(e.target.value)}
                    allowClear
                    autoCapitalize="none"
                    autoCorrect="off"
                    spellCheck={false}
                    style={{ width: '150px' }}
                  />
                  <Tooltip title="Negate tag filter (exclude instead of include)">
                    <Switch
                      checked={negateTagFilter}
                      onChange={setNegateTagFilter}
                      checkedChildren="NOT"
                      unCheckedChildren="IS"
                      size="small"
                    />
                  </Tooltip>
                  <Button
                    size="small"
                    onClick={loadWorkspacesByTagAdvanced}
                    loading={loading}
                    disabled={!tagFilter}
                  >
                    Apply Tag Filter
                  </Button>
                </Space>
              </Col>
            </Row>
          </Card>

          {/* Workspace Selection */}
          <Card size="small" style={{ marginBottom: '16px' }}>
            <Row gutter={[16, 16]}>
              <Col xs={24} md={12}>
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Text strong>Select Workspace:</Text>
                  <Select
                    placeholder="Choose a workspace..."
                    value={selectedWorkspaceForDetails}
                    onChange={handleWorkspaceSelection}
                    style={{ width: '100%' }}
                    showSearch
                    filterOption={(input, option) =>
                      option?.children?.toLowerCase().indexOf(input.toLowerCase()) >= 0
                    }
                    loading={loading}
                  >
                    {filteredWorkspaces.map(ws => (
                      <Option key={ws.name} value={ws.name}>
                        {ws.name}
                      </Option>
                    ))}
                  </Select>
                </Space>
              </Col>
              <Col xs={24} md={12}>
                {selectedWorkspaceForDetails && (
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Text strong>Actions:</Text>
                    <Space>
                      <Button
                        icon={<LockOutlined />}
                        onClick={() => lockWorkspace(selectedWorkspaceForDetails)}
                        loading={loading}
                        size="small"
                        disabled={selectedWorkspaceDetails?.locked}
                      >
                        Lock
                      </Button>
                      <Button
                        icon={<UnlockOutlined />}
                        onClick={() => unlockWorkspace(selectedWorkspaceForDetails)}
                        loading={loading}
                        size="small"
                        disabled={!selectedWorkspaceDetails?.locked}
                      >
                        Unlock
                      </Button>
                      <Select
                        placeholder="Set TF Version..."
                        style={{ width: '150px' }}
                        size="small"
                        onSelect={(version) => setWorkspaceVersion(selectedWorkspaceForDetails, version)}
                        loading={loading}
                      >
                        <Option value="1.5.0">1.5.0</Option>
                        <Option value="1.4.6">1.4.6</Option>
                        <Option value="1.3.10">1.3.10</Option>
                        <Option value="1.2.9">1.2.9</Option>
                      </Select>
                    </Space>
                  </Space>
                )}
              </Col>
            </Row>
          </Card>

          {/* Workspace Details */}
          {selectedWorkspaceDetails && (
            <Card 
              title={
                <span>
                  Workspace Details: <Text copyable>{selectedWorkspaceDetails.name}</Text>
                </span>
              } 
              size="small"
            >
              <Row gutter={[16, 16]}>
                <Col xs={24} md={12}>
                  <Card size="small" title="Basic Information" type="inner">
                    <p><strong>Name:</strong> <Text copyable>{selectedWorkspaceDetails.name}</Text></p>
                    <p><strong>ID:</strong> <Text code copyable>{selectedWorkspaceDetails.id}</Text></p>
                    <p><strong>Organization:</strong> {selectedWorkspaceDetails.organization}</p>
                    <p><strong>Environment:</strong> <Tag color="blue">{extractEnvironmentFromName(selectedWorkspaceDetails.name)}</Tag></p>
                  </Card>
                </Col>
                <Col xs={24} md={12}>
                  <Card size="small" title="Configuration" type="inner">
                    <p><strong>Workspace Type:</strong> <Tag color="purple">{selectedWorkspaceDetails.workspace_type || 'Unknown'}</Tag></p>
                    <p><strong>Execution Mode:</strong> <Tag color="cyan">{selectedWorkspaceDetails.execution_mode || 'Unknown'}</Tag></p>
                    <p><strong>Terraform Version:</strong> <Text code copyable>{selectedWorkspaceDetails.terraform_version || 'Not set'}</Text></p>
                    <p><strong>Auto Apply:</strong> <Tag color={selectedWorkspaceDetails.auto_apply ? 'green' : 'red'}>{selectedWorkspaceDetails.auto_apply ? 'Enabled' : 'Disabled'}</Tag></p>
                    <p><strong>Locked:</strong> <Tag color={selectedWorkspaceDetails.locked ? 'orange' : 'green'}>{selectedWorkspaceDetails.locked ? 'Locked' : 'Unlocked'}</Tag></p>
                  </Card>
                </Col>
              </Row>
              {selectedWorkspaceDetails.vcs_connection && (
                <Row gutter={[16, 16]} style={{ marginTop: '16px' }}>
                  <Col span={24}>
                    <Card size="small" title="VCS Connection" type="inner">
                      <Row gutter={[16, 8]}>
                        <Col span={12}>
                          <p><strong>Repository:</strong> <Text code copyable>{selectedWorkspaceDetails.vcs_connection.repository}</Text></p>
                          <p><strong>Branch:</strong> <Text code copyable>{selectedWorkspaceDetails.vcs_connection.branch}</Text></p>
                        </Col>
                        <Col span={12}>
                          <p><strong>Working Directory:</strong> <Text code copyable>{selectedWorkspaceDetails.vcs_connection.working_directory || '/'}</Text></p>
                          {selectedWorkspaceDetails.vcs_connection.webhook_url && (
                            <p><strong>Webhook URL:</strong> <Text code copyable style={{ wordBreak: 'break-all' }}>{selectedWorkspaceDetails.vcs_connection.webhook_url}</Text></p>
                          )}
                        </Col>
                      </Row>
                    </Card>
                  </Col>
                </Row>
              )}
              {selectedWorkspaceDetails.tag_names && selectedWorkspaceDetails.tag_names.length > 0 && (
                <Row gutter={[16, 16]} style={{ marginTop: '16px' }}>
                  <Col span={24}>
                    <Card size="small" title="Tags" type="inner">
                      <Space wrap>
                        {selectedWorkspaceDetails.tag_names.map((tag: string, index: number) => (
                          <Tag key={`${tag}-${index}`} color="blue">{tag}</Tag>
                        ))}
                      </Space>
                    </Card>
                  </Col>
                </Row>
              )}
            </Card>
          )}
        </div>
      ),
    },
    {
      key: 'runs',
      label: (
        <span>
          <PlayCircleOutlined />
          Runs
          {loading && <Spin size="small" style={{ marginLeft: '8px' }} />}
        </span>
      ),
      disabled: loading || workspaces.length === 0,
      children: (
        <div>
          <Row justify="space-between" style={{ marginBottom: '16px' }}>
            <Col>
              <Space>
                <Select
                  placeholder="Select workspace for runs..."
                  value={selectedWorkspace}
                  onChange={(value) => {
                    setSelectedWorkspace(value);
                    loadRuns(value);
                  }}
                  style={{ width: '400px' }}
                  showSearch
                  filterOption={(input, option) =>
                    option?.children?.toLowerCase().indexOf(input.toLowerCase()) >= 0
                  }
                >
                  {workspaces.map(ws => (
                    <Option key={ws.id} value={ws.id}>
                      {ws.name}
                    </Option>
                  ))}
                </Select>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={() => loadRuns()}
                  loading={loading}
                  disabled={!selectedWorkspace}
                >
                  Refresh
                </Button>
                <Button
                  icon={<DeleteOutlined />}
                  onClick={() => discardOldRuns(24)}
                  loading={loading}
                  type="primary"
                  danger
                  disabled={!selectedWorkspace}
                >
                  Discard Old Runs
                </Button>
              </Space>
            </Col>
            <Col>
              <Text type="secondary">
                {selectedWorkspace ? `${getFilteredRuns().length} of ${runs.length} runs` : 'Select workspace to view runs'}
              </Text>
            </Col>
          </Row>
          
          {/* Run Filter Controls */}
          <Card size="small" style={{ marginBottom: '16px' }}>
            <Row gutter={[16, 16]}>
              <Col xs={24} sm={12} md={8}>
                <Select
                  placeholder="Filter by status..."
                  value={runStatusFilter}
                  onChange={setRunStatusFilter}
                  allowClear
                  style={{ width: '100%' }}
                >
                  <Option value="">All Statuses</Option>
                  {getUniqueRunStatuses().map(status => (
                    <Option key={status} value={status}>
                      {status}
                    </Option>
                  ))}
                </Select>
              </Col>
              <Col xs={24} sm={12} md={8}>
                <Select
                  placeholder="Filter by date..."
                  value={runDateFilter}
                  onChange={setRunDateFilter}
                  allowClear
                  style={{ width: '100%' }}
                >
                  <Option value="">All Time</Option>
                  <Option value="24h">Last 24 Hours</Option>
                  <Option value="7d">Last 7 Days</Option>
                  <Option value="30d">Last 30 Days</Option>
                </Select>
              </Col>
              <Col xs={24} sm={12} md={8}>
                <Space>
                  <Button
                    size="small"
                    onClick={() => discardOldRuns(168)} // 7 days
                    loading={loading}
                    danger
                  >
                    Discard 7d+ Old
                  </Button>
                  <Button
                    size="small"
                    onClick={() => discardOldRuns(720)} // 30 days
                    loading={loading}
                    danger
                  >
                    Discard 30d+ Old
                  </Button>
                </Space>
              </Col>
            </Row>
          </Card>

          {/* Bulk Operations for Runs */}
          {selectedRuns.length > 0 && (
            <Card size="small" style={{ marginBottom: '16px', backgroundColor: '#fff2e8' }}>
              <Row justify="space-between" align="middle">
                <Col>
                  <Text strong>
                    {selectedRuns.length} run{selectedRuns.length > 1 ? 's' : ''} selected
                  </Text>
                </Col>
                <Col>
                  <Space>
                    <Button
                      icon={<DeleteOutlined />}
                      onClick={bulkDiscardRuns}
                      loading={loading}
                      size="small"
                      danger
                    >
                      Discard Selected
                    </Button>
                    <Button
                      type="text"
                      onClick={() => setSelectedRuns([])}
                      size="small"
                    >
                      Clear Selection
                    </Button>
                  </Space>
                </Col>
              </Row>
            </Card>
          )}

          <Table
            dataSource={getFilteredRuns()}
            columns={runColumns}
            rowKey="id"
            loading={loading}
            rowSelection={{
              type: 'checkbox',
              selectedRowKeys: selectedRuns,
              onChange: (selectedRowKeys) => setSelectedRuns(selectedRowKeys as string[]),
              selections: [
                Table.SELECTION_ALL,
                Table.SELECTION_INVERT,
                Table.SELECTION_NONE,
                {
                  key: 'select-failed',
                  text: 'Select Failed',
                  onSelect: (changableRowKeys) => {
                    const failedRuns = getFilteredRuns()
                      .filter(run => run.status === 'errored' || run.status === 'canceled')
                      .map(run => run.id);
                    setSelectedRuns(failedRuns);
                  },
                },
                {
                  key: 'select-old',
                  text: 'Select Old (7d+)',
                  onSelect: (changableRowKeys) => {
                    const sevenDaysAgo = new Date();
                    sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);
                    const oldRuns = getFilteredRuns()
                      .filter(run => new Date(run.created_at) < sevenDaysAgo)
                      .map(run => run.id);
                    setSelectedRuns(oldRuns);
                  },
                },
              ],
            }}
            pagination={{ 
              pageSize: pageSize,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} runs`,
              pageSizeOptions: ['10', '20', '50', '100'],
              onShowSizeChange: (current, size) => setPageSize(size)
            }}
            size="small"
          />
          
          {!selectedWorkspace && (
            <div style={{ textAlign: 'center', padding: '50px' }}>
              <PlayCircleOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
              <div style={{ marginTop: '16px', color: '#999' }}>
                Select a workspace to view its runs
              </div>
            </div>
          )}
        </div>
      ),
    },
    {
      key: 'variables',
      label: (
        <span>
          <EditOutlined />
          Variables
          {loading && <Spin size="small" style={{ marginLeft: '8px' }} />}
        </span>
      ),
      disabled: loading || workspaces.length === 0,
      children: (
        <div>
          <Row justify="space-between" style={{ marginBottom: '16px' }}>
            <Col>
              <Space>
                <Button.Group>
                  <Button
                    type={variableViewMode === 'workspace' ? 'primary' : 'default'}
                    onClick={() => setVariableViewMode('workspace')}
                  >
                    Workspace Variables
                  </Button>
                  <Button
                    type={variableViewMode === 'variable-set' ? 'primary' : 'default'}
                    onClick={() => setVariableViewMode('variable-set')}
                  >
                    Variable Sets
                  </Button>
                </Button.Group>
                
                {variableViewMode === 'workspace' ? (
                  <Space>
                    <Select
                      placeholder="Select workspace..."
                      value={selectedWorkspaceForVars}
                      onChange={(value) => {
                        setSelectedWorkspaceForVars(value);
                        if (value) {
                          loadWorkspaceVariables(value);
                        }
                      }}
                      style={{ width: '400px' }}
                      showSearch
                      filterOption={(input, option) =>
                        option?.children?.toLowerCase().indexOf(input.toLowerCase()) >= 0
                      }
                    >
                      {workspaces.map(ws => (
                        <Option key={ws.id} value={ws.id}>
                          {ws.name}
                        </Option>
                      ))}
                    </Select>
                    <Tooltip title="Include variables from variable sets assigned to this workspace">
                      <Switch
                        checked={includeVariableSets}
                        onChange={setIncludeVariableSets}
                        checkedChildren="Include Sets"
                        unCheckedChildren="Workspace Only"
                        size="small"
                      />
                    </Tooltip>
                    <Button
                      icon={<ReloadOutlined />}
                      onClick={() => {
                        if (selectedWorkspaceForVars) {
                          loadWorkspaceVariables(selectedWorkspaceForVars);
                        }
                      }}
                      loading={loading}
                      disabled={!selectedWorkspaceForVars}
                    >
                      Refresh Variables
                    </Button>
                  </Space>
                ) : (
                  <Space>
                    <Select
                      placeholder="Select variable set..."
                      value={selectedVariableSet}
                      onChange={(value) => {
                        setSelectedVariableSet(value);
                        if (value) {
                          loadVariableSetDetails(value);
                        }
                      }}
                      style={{ width: '400px' }}
                      showSearch
                      filterOption={(input, option) =>
                        option?.children?.toLowerCase().indexOf(input.toLowerCase()) >= 0
                      }
                      notFoundContent={variableSets.length === 0 ? 'No variable sets found' : 'No matching variable sets'}
                    >
                      {variableSets.map(vs => (
                        <Option key={vs.id} value={vs.name}>
                          {vs.name}
                        </Option>
                      ))}
                    </Select>
                    <Button
                      icon={<ReloadOutlined />}
                      onClick={() => {
                        if (selectedVariableSet) {
                          loadVariableSetDetails(selectedVariableSet);
                        }
                      }}
                      loading={loading}
                      disabled={!selectedVariableSet}
                    >
                      Refresh Variables
                    </Button>
                    <Button
                      icon={<ReloadOutlined />}
                      onClick={loadVariableSets}
                      loading={loading}
                      size="small"
                      title="Refresh variable sets list"
                    >
                      Refresh Sets
                    </Button>
                  </Space>
                )}
              </Space>
            </Col>
          </Row>
          
          {/* Display associated workspaces for variable-set mode */}
          {variableViewMode === 'variable-set' && variableSetDetails && variableSetDetails.workspaces && (
            <Row gutter={[16, 16]} style={{ marginBottom: '16px' }}>
              <Col span={24}>
                <Card title="Associated Workspaces" size="small">
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
                    {variableSetDetails.workspaces.map((workspace: any) => (
                      <Tag key={workspace.id} color="blue">
                        {workspace.name}
                      </Tag>
                    ))}
                  </div>
                </Card>
              </Col>
            </Row>
          )}
          
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={12}>
              <Card title="Terraform Variables" size="small">
                <Table
                  dataSource={workspaceVariables.filter(v => v.category === 'terraform')}
                  columns={[
                    {
                      title: 'Key',
                      dataIndex: 'key',
                      key: 'key',
                      render: (text: string) => <Text code>{text}</Text>,
                    },
                    {
                      title: 'Value',
                      dataIndex: 'value',
                      key: 'value',
                      render: (text: string, record: any) => (
                        record.sensitive ? <Text type="secondary">***SENSITIVE***</Text> : <Text>{text || 'N/A'}</Text>
                      ),
                    },
                    {
                      title: 'HCL',
                      dataIndex: 'hcl',
                      key: 'hcl',
                      render: (hcl: boolean) => (
                        hcl ? <Tag color="blue">HCL</Tag> : <Tag>String</Tag>
                      ),
                    },
                    {
                      title: 'Source',
                      dataIndex: 'source',
                      key: 'source',
                      render: (source: string) => (
                        <Tag color={source === 'workspace' ? 'green' : source === 'variable-set' ? 'blue' : 'orange'}>
                          {source || 'workspace'}
                        </Tag>
                      ),
                    },
                  ]}
                  rowKey={(record) => `${record.key}-${record.source || 'workspace'}`}
                  loading={loading}
                  pagination={false}
                  size="small"
                  scroll={{ y: 300 }}
                />
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card title="Environment Variables" size="small">
                <Table
                  dataSource={workspaceVariables.filter(v => v.category === 'env')}
                  columns={[
                    {
                      title: 'Key',
                      dataIndex: 'key',
                      key: 'key',
                      render: (text: string) => <Text code>{text}</Text>,
                    },
                    {
                      title: 'Value',
                      dataIndex: 'value',
                      key: 'value',
                      render: (text: string, record: any) => (
                        record.sensitive ? <Text type="secondary">***SENSITIVE***</Text> : <Text>{text || 'N/A'}</Text>
                      ),
                    },
                    {
                      title: 'Source',
                      dataIndex: 'source',
                      key: 'source',
                      render: (source: string) => (
                        <Tag color={source === 'workspace' ? 'green' : source === 'variable-set' ? 'blue' : 'orange'}>
                          {source || 'workspace'}
                        </Tag>
                      ),
                    },
                  ]}
                  rowKey={(record) => `${record.key}-${record.source || 'workspace'}`}
                  loading={loading}
                  pagination={false}
                  size="small"
                  scroll={{ y: 300 }}
                />
              </Card>
            </Col>
          </Row>
          
          {variableViewMode === 'workspace' && !selectedWorkspaceForVars && (
            <div style={{ textAlign: 'center', padding: '50px' }}>
              <EditOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
              <div style={{ marginTop: '16px', color: '#999' }}>
                Select a workspace to view its variables
              </div>
            </div>
          )}
          
          {variableViewMode === 'variable-set' && !selectedVariableSet && (
            <div style={{ textAlign: 'center', padding: '50px' }}>
              <EditOutlined style={{ fontSize: '48px', color: '#d9d9d9' }} />
              <div style={{ marginTop: '16px', color: '#999' }}>
                Select a variable set to view its variables
              </div>
            </div>
          )}
        </div>
      ),
    },
    {
      key: 'versions',
      label: (
        <span>
          <FileTextOutlined />
          Versions
          {loading && <Spin size="small" style={{ marginLeft: '8px' }} />}
        </span>
      ),
      disabled: loading || workspaces.length === 0,
      children: (
        <div>
          <Row justify="space-between" style={{ marginBottom: '16px' }}>
            <Col>
              <Space>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={loadTerraformVersions}
                  loading={loading}
                >
                  Refresh Versions
                </Button>
                <Button
                  icon={<ExclamationCircleOutlined />}
                  onClick={checkDeprecatedVersions}
                  loading={loading}
                  type="primary"
                >
                  Check Deprecated Versions
                </Button>
              </Space>
            </Col>
          </Row>
          
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={12}>
              <Card title="Terraform Versions" size="small">
                <Table
                  dataSource={terraformVersions}
                  columns={versionColumns}
                  rowKey="version"
                  loading={loading}
                  pagination={false}
                  size="small"
                  scroll={{ y: 300 }}
                />
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card title="Version Summary" size="small">
                {versionSummary && (
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Row justify="space-between">
                      <Col>
                        <Text strong>Total Workspaces:</Text>
                      </Col>
                      <Col>
                        <Text>{versionSummary.totalWorkspaces}</Text>
                      </Col>
                    </Row>
                    <Row justify="space-between">
                      <Col>
                        <Text strong>Deprecated Versions:</Text>
                      </Col>
                      <Col>
                        <Text type="danger">{versionSummary.deprecatedCount}</Text>
                      </Col>
                    </Row>
                    <Row justify="space-between">
                      <Col>
                        <Text strong>Current Versions:</Text>
                      </Col>
                      <Col>
                        <Text type="success">{versionSummary.currentCount}</Text>
                      </Col>
                    </Row>
                    <Divider />
                    <Text strong>Most Used Versions:</Text>
                    {versionSummary.topVersions.map((version, index) => (
                      <Row key={version.version} justify="space-between">
                        <Col>
                          <Text code>{version.version}</Text>
                        </Col>
                        <Col>
                          <Tag color={version.isDeprecated ? 'red' : 'green'}>
                            {version.count} workspaces
                          </Tag>
                        </Col>
                      </Row>
                    ))}
                  </Space>
                )}
              </Card>
            </Col>
          </Row>
        </div>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Row justify="space-between" style={{ marginBottom: '24px' }}>
        <Col>
          <Title level={2}>
            <CloudServerOutlined style={{ marginRight: '8px' }} />
            Terraform Enterprise
            <Tag color="orange" style={{ marginLeft: '8px' }}>BETA</Tag>
          </Title>
          <Text type="secondary">Manage TFE workspaces, runs, and versions</Text>
        </Col>
        <Col>
          <Space>
            <Button
              icon={<SettingOutlined />}
              onClick={() => setShowConfigModal(true)}
            >
              Configure
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

      <Card>
        <Space direction="vertical" style={{ width: '100%', marginBottom: '16px' }}>
          <Row justify="space-between">
            <Col>
              <Text strong>Endpoint:</Text> <Text code>{config.endpoint}</Text>
            </Col>
            <Col>
              <Text strong>Organization:</Text> <Text code>{config.organization}</Text>
            </Col>
          </Row>
        </Space>

        <Tabs
          activeKey={activeTab}
          onChange={(key) => setActiveTab(key as any)}
          items={tabItems}
        />
      </Card>

      {/* Configuration Modal */}
      <Modal
        title="TFE Configuration"
        open={showConfigModal}
        onCancel={() => setShowConfigModal(false)}
        footer={null}
        width={600}
      >
        <Form
          form={configForm}
          layout="vertical"
          onFinish={handleConfigSave}
          initialValues={config}
        >
          <Alert
            message="TFE Configuration"
            description="Configure your Terraform Enterprise endpoint and organization. Changes will be applied immediately."
            type="info"
            showIcon
            style={{ marginBottom: '16px' }}
          />
          
          <Form.Item
            name="endpoint"
            label="TFE Endpoint"
            rules={[{ required: true, message: 'Please enter the TFE endpoint' }]}
          >
            <Input
              placeholder="e.g., tfe.doctolib.net or app.terraform.io"
              prefix={<CloudServerOutlined />}
            />
          </Form.Item>
          
          <Form.Item
            name="organization"
            label="Organization"
            rules={[{ required: true, message: 'Please enter the organization name' }]}
          >
            <Input
              placeholder="e.g., doctolib"
              prefix={<DatabaseOutlined />}
            />
          </Form.Item>
          
          <Form.Item
            name="token"
            label="TFE Token (Optional)"
            help="Token can also be set via TFE_TOKEN environment variable"
          >
            <Input.Password
              placeholder="Enter TFE token (optional)"
              prefix={<LockOutlined />}
            />
          </Form.Item>
          
          <Form.Item style={{ marginBottom: 0 }}>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                Save Configuration
              </Button>
              <Button onClick={() => setShowConfigModal(false)}>
                Cancel
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>


      {/* Run Logs Modal */}
      <Modal
        title={`Run Logs - ${selectedRunForLogs}`}
        open={runLogsVisible}
        onCancel={() => {
          setRunLogsVisible(false);
          setSelectedRunForLogs('');
          setRunLogs('');
        }}
        footer={[
          <Button key="copy" icon={<CopyOutlined />} onClick={() => navigator.clipboard.writeText(runLogs)}>
            Copy Logs
          </Button>,
          <Button key="close" onClick={() => {
            setRunLogsVisible(false);
            setSelectedRunForLogs('');
            setRunLogs('');
          }}>
            Close
          </Button>
        ]}
        width={1000}
        style={{ top: 20 }}
      >
        {runLogsLoading ? (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <Spin size="large" />
            <div style={{ marginTop: '16px' }}>
              <Text>Loading run logs...</Text>
            </div>
          </div>
        ) : (
          <div style={{ maxHeight: '60vh', overflow: 'auto' }}>
            <pre style={{ 
              fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
              fontSize: '12px',
              lineHeight: '1.4',
              backgroundColor: '#f6f8fa',
              padding: '16px',
              borderRadius: '4px',
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word'
            }}>
              {runLogs || 'No logs available'}
            </pre>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default TFE;