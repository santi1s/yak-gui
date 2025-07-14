import React, { useState, useEffect } from 'react';
import { 
  ArrowPathIcon, 
  PlayIcon, 
  PauseIcon, 
  DocumentTextIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  ServerIcon,
  CogIcon,
  ComputerDesktopIcon,
  FolderIcon
} from '@heroicons/react/24/outline';
import Rollouts from './Rollouts';
import Secrets from './Secrets';

// Types matching the Go backend
interface ArgoApp {
  AppName: string;
  Health: string;
  Sync: string;
  Suspended: boolean;
  SyncLoop: string;
  Conditions: string[];
}

interface ArgoConfig {
  server: string;
  project: string;
  username?: string;
  password?: string;
}

// Declare global functions for Wails
declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetArgoApps: (config: ArgoConfig) => Promise<ArgoApp[]>;
          SyncArgoApp: (config: ArgoConfig, appName: string, prune: boolean, dryRun: boolean) => Promise<void>;
          RefreshArgoApp: (config: ArgoConfig, appName: string) => Promise<void>;
          SuspendArgoApp: (config: ArgoConfig, appName: string) => Promise<void>;
          UnsuspendArgoApp: (config: ArgoConfig, appName: string) => Promise<void>;
          GetArgoCDServerFromProfile: () => Promise<string>;
          GetCurrentAWSProfile: () => Promise<string>;
          SetAWSProfile: (profile: string) => Promise<void>;
          GetKubeconfig: () => Promise<string>;
          SetKubeconfig: (path: string) => Promise<void>;
          GetEnvironmentVariables: () => Promise<Record<string, string>>;
          TestSimpleArray: () => Promise<string[]>;
          TestSimpleApps: () => Promise<ArgoApp[]>;
          LoginToArgoCD: (config: ArgoConfig) => Promise<void>;
          // Rollouts functions
          GetRollouts: (config: any) => Promise<any[]>;
          GetRolloutStatus: (config: any, rolloutName: string) => Promise<any>;
          PromoteRollout: (config: any, rolloutName: string, full: boolean) => Promise<void>;
          PauseRollout: (config: any, rolloutName: string) => Promise<void>;
          AbortRollout: (config: any, rolloutName: string) => Promise<void>;
          RestartRollout: (config: any, rolloutName: string) => Promise<void>;
          SetRolloutImage: (config: any, rolloutName: string, image: string, container: string) => Promise<void>;
          // Secret functions
          GetSecrets: (config: any, path: string) => Promise<any[]>;
          GetSecretData: (config: any, path: string, version: number) => Promise<any>;
          CreateSecret: (config: any, path: string, owner: string, usage: string, source: string, data: Record<string, string>) => Promise<void>;
          UpdateSecret: (config: any, path: string, data: Record<string, string>) => Promise<void>;
          DeleteSecret: (config: any, path: string, version: number) => Promise<void>;
        };
      };
    };
  }
}

const StatusBadge: React.FC<{ status: string, type: 'health' | 'sync' | 'syncLoop' }> = ({ status, type }) => {
  const getStatusClass = () => {
    if (type === 'health') {
      switch (status.toLowerCase()) {
        case 'healthy': return 'status-healthy';
        case 'progressing': return 'status-progressing';
        case 'degraded': return 'status-degraded';
        case 'suspended': return 'status-suspended';
        case 'missing': return 'status-missing';
        default: return 'status-unknown';
      }
    } else if (type === 'sync') {
      switch (status.toLowerCase()) {
        case 'synced': return 'status-synced';
        case 'outofsync': return 'status-outofsync';
        default: return 'status-unknown';
      }
    } else if (type === 'syncLoop') {
      switch (status.toLowerCase()) {
        case 'critical': return 'sync-loop-critical';
        case 'warning': return 'sync-loop-warning';
        case 'possible': return 'sync-loop-possible';
        case 'failed': return 'sync-loop-failed';
        default: return 'bg-gray-600 text-white';
      }
    }
    return 'bg-gray-600 text-white';
  };

  return (
    <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusClass()}`}>
      {status}
    </span>
  );
};

const AppCard: React.FC<{ 
  app: ArgoApp; 
  config: ArgoConfig;
  onAction: () => void;
}> = ({ app, config, onAction }) => {
  const [isLoading, setIsLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const handleSync = async () => {
    setActionLoading('sync');
    try {
      await window.go.main.App.SyncArgoApp(config, app?.AppName || '', false, false);
      onAction();
    } catch (error) {
      console.error('Failed to sync app:', error);
    } finally {
      setActionLoading(null);
    }
  };

  const handleRefresh = async () => {
    setActionLoading('refresh');
    try {
      await window.go.main.App.RefreshArgoApp(config, app?.AppName || '');
      onAction();
    } catch (error) {
      console.error('Failed to refresh app:', error);
    } finally {
      setActionLoading(null);
    }
  };

  const handleSuspend = async () => {
    setActionLoading('suspend');
    try {
      if (app?.Suspended) {
        await window.go.main.App.UnsuspendArgoApp(config, app?.AppName || '');
      } else {
        await window.go.main.App.SuspendArgoApp(config, app?.AppName || '');
      }
      onAction();
    } catch (error) {
      console.error('Failed to suspend/unsuspend app:', error);
    } finally {
      setActionLoading(null);
    }
  };

  const getHealthIcon = () => {
    const health = (app?.Health || '').toLowerCase();
    switch (health) {
      case 'healthy':
        return <CheckCircleIcon className="w-5 h-5 text-green-500" />;
      case 'degraded':
        return <XCircleIcon className="w-5 h-5 text-red-500" />;
      case 'progressing':
        return <ClockIcon className="w-5 h-5 text-yellow-500" />;
      default:
        return <ExclamationTriangleIcon className="w-5 h-5 text-gray-500" />;
    }
  };

  return (
    <div className="bg-slate-800 rounded-lg p-6 border border-slate-700 hover:border-slate-600 transition-colors">
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center space-x-3">
          {getHealthIcon()}
          <h3 className="text-lg font-semibold text-white">{app?.AppName || 'Unknown'}</h3>
        </div>
        {app?.Suspended && (
          <span className="px-2 py-1 bg-yellow-600 text-white text-xs rounded-full">
            SUSPENDED
          </span>
        )}
      </div>

      <div className="space-y-3 mb-4">
        <div className="flex items-center justify-between">
          <span className="text-sm text-slate-400">Health:</span>
          <StatusBadge status={app?.Health || 'Unknown'} type="health" />
        </div>
        
        <div className="flex items-center justify-between">
          <span className="text-sm text-slate-400">Sync:</span>
          <StatusBadge status={app?.Sync || 'Unknown'} type="sync" />
        </div>

        {app?.SyncLoop && app.SyncLoop !== 'No' && (
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Sync Loop:</span>
            <StatusBadge status={app.SyncLoop} type="syncLoop" />
          </div>
        )}

        {app?.Conditions && app.Conditions.length > 0 && (
          <div className="flex items-start justify-between">
            <span className="text-sm text-slate-400">Conditions:</span>
            <div className="flex flex-wrap gap-1 max-w-48">
              {app.Conditions.map((condition, idx) => (
                <span key={idx} className="px-2 py-1 bg-blue-600 text-white text-xs rounded">
                  {condition}
                </span>
              ))}
            </div>
          </div>
        )}
      </div>

      <div className="flex space-x-2">
        <button
          onClick={handleSync}
          disabled={actionLoading !== null}
          className="flex-1 flex items-center justify-center space-x-2 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 disabled:opacity-50 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
        >
          {actionLoading === 'sync' ? (
            <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
          ) : (
            <ArrowPathIcon className="w-4 h-4" />
          )}
          <span>Sync</span>
        </button>

        <button
          onClick={handleRefresh}
          disabled={actionLoading !== null}
          className="flex items-center justify-center bg-slate-600 hover:bg-slate-700 disabled:bg-slate-800 disabled:opacity-50 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
        >
          {actionLoading === 'refresh' ? (
            <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
          ) : (
            <DocumentTextIcon className="w-4 h-4" />
          )}
        </button>

        <button
          onClick={handleSuspend}
          disabled={actionLoading !== null}
          className={`flex items-center justify-center px-3 py-2 rounded-md text-sm font-medium transition-colors ${
            app?.Suspended 
              ? 'bg-green-600 hover:bg-green-700 disabled:bg-green-800' 
              : 'bg-yellow-600 hover:bg-yellow-700 disabled:bg-yellow-800'
          } disabled:opacity-50 text-white`}
        >
          {actionLoading === 'suspend' ? (
            <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
          ) : app?.Suspended ? (
            <PlayIcon className="w-4 h-4" />
          ) : (
            <PauseIcon className="w-4 h-4" />
          )}
        </button>
      </div>
    </div>
  );
};

// Environment Configuration Component
const EnvironmentConfig: React.FC = () => {
  const [envVars, setEnvVars] = useState<Record<string, string>>({});
  const [awsProfile, setAwsProfile] = useState('');
  const [kubeconfig, setKubeconfig] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const loadEnvironmentVariables = async () => {
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const vars = await window.go.main.App.GetEnvironmentVariables();
        setEnvVars(vars);
        setAwsProfile(vars.AWS_PROFILE || '');
        setKubeconfig(vars.KUBECONFIG || '');
      }
    } catch (error) {
      console.error('Failed to load environment variables:', error);
      setError('Failed to load environment variables');
    }
  };

  useEffect(() => {
    loadEnvironmentVariables();
  }, []);

  const handleSetAWSProfile = async () => {
    if (!awsProfile.trim()) {
      setError('AWS Profile cannot be empty');
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      await window.go.main.App.SetAWSProfile(awsProfile.trim());
      setSuccess('AWS Profile set successfully');
      await loadEnvironmentVariables();
    } catch (error) {
      setError(`Failed to set AWS Profile: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const handleSetKubeconfig = async () => {
    if (!kubeconfig.trim()) {
      setError('Kubeconfig path cannot be empty');
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);
    
    try {
      await window.go.main.App.SetKubeconfig(kubeconfig.trim());
      setSuccess('Kubeconfig set successfully');
      await loadEnvironmentVariables();
    } catch (error) {
      setError(`Failed to set Kubeconfig: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-900 text-white p-6">
      <div className="max-w-4xl mx-auto">
        <div className="flex items-center space-x-3 mb-6">
          <ComputerDesktopIcon className="w-8 h-8 text-green-500" />
          <div>
            <h1 className="text-2xl font-bold">Environment Configuration</h1>
            <p className="text-slate-400">Set environment variables for AWS and Kubernetes</p>
          </div>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-red-900 border border-red-700 rounded-lg">
            <div className="flex items-center space-x-2">
              <XCircleIcon className="w-5 h-5 text-red-400" />
              <span className="text-red-100">{error}</span>
            </div>
          </div>
        )}

        {success && (
          <div className="mb-4 p-4 bg-green-900 border border-green-700 rounded-lg">
            <div className="flex items-center space-x-2">
              <CheckCircleIcon className="w-5 h-5 text-green-400" />
              <span className="text-green-100">{success}</span>
            </div>
          </div>
        )}

        <div className="grid gap-6">
          {/* AWS Profile Configuration */}
          <div className="bg-slate-800 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
              <ServerIcon className="w-5 h-5 text-blue-400" />
              <span>AWS Profile</span>
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Current AWS Profile: <span className="text-blue-400">{envVars.AWS_PROFILE || 'Not set'}</span>
                </label>
                <div className="flex space-x-3">
                  <input
                    type="text"
                    value={awsProfile}
                    onChange={(e) => setAwsProfile(e.target.value)}
                    placeholder="staging-aws-fr-par-1"
                    className="flex-1 px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <button
                    onClick={handleSetAWSProfile}
                    disabled={loading}
                    className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 disabled:opacity-50 text-white rounded-md font-medium transition-colors"
                  >
                    {loading ? 'Setting...' : 'Set Profile'}
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* Kubeconfig Configuration */}
          <div className="bg-slate-800 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
              <FolderIcon className="w-5 h-5 text-purple-400" />
              <span>Kubernetes Configuration</span>
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Current KUBECONFIG: <span className="text-purple-400 text-xs font-mono break-all">{envVars.KUBECONFIG || 'Not set'}</span>
                </label>
                <div className="flex space-x-3">
                  <input
                    type="text"
                    value={kubeconfig}
                    onChange={(e) => setKubeconfig(e.target.value)}
                    placeholder="/path/to/kubeconfig"
                    className="flex-1 px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
                  />
                  <button
                    onClick={handleSetKubeconfig}
                    disabled={loading}
                    className="px-4 py-2 bg-purple-600 hover:bg-purple-700 disabled:bg-purple-800 disabled:opacity-50 text-white rounded-md font-medium transition-colors"
                  >
                    {loading ? 'Setting...' : 'Set Config'}
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* Environment Variables Display */}
          <div className="bg-slate-800 rounded-lg p-6">
            <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
              <CogIcon className="w-5 h-5 text-green-400" />
              <span>Current Environment</span>
            </h2>
            <div className="space-y-2">
              {Object.entries(envVars).map(([key, value]) => (
                <div key={key} className="flex">
                  <span className="w-32 text-slate-400 font-mono text-sm">{key}:</span>
                  <span className="text-white font-mono text-sm break-all">{value || 'Not set'}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

const App: React.FC = () => {
  const [apps, setApps] = useState<ArgoApp[]>([]);
  const [loading, setLoading] = useState(false);
  const [config, setConfig] = useState<ArgoConfig>({
    server: '',
    project: 'main',
    username: '',
    password: ''
  });
  const [error, setError] = useState<string | null>(null);
  const [showConfig, setShowConfig] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [awsProfile, setAwsProfile] = useState<string>('');
  const [isLoggingIn, setIsLoggingIn] = useState(false);
  const [activeTab, setActiveTab] = useState<'argocd' | 'rollouts' | 'secrets' | 'environment'>('argocd');

  const loadAWSProfile = async () => {
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const profile = await window.go.main.App.GetCurrentAWSProfile();
        setAwsProfile(profile);
        
        if (profile && !config.server) {
          // Auto-detect ArgoCD server from AWS profile
          const server = await window.go.main.App.GetArgoCDServerFromProfile();
          setConfig(prev => ({ ...prev, server }));
        }
      }
    } catch (error) {
      console.error('Failed to load AWS profile:', error);
      // Don't set error state for this, it's optional
    }
  };

  const handleLogin = async () => {
    setIsLoggingIn(true);
    setError(null);
    try {
      await window.go.main.App.LoginToArgoCD(config);
      // After successful login, try to load apps again
      setTimeout(() => loadApps(), 1000); // Small delay to ensure login is complete
    } catch (error) {
      setError(`Login failed: ${error instanceof Error ? error.message : String(error)}`);
    } finally {
      setIsLoggingIn(false);
    }
  };

  const loadApps = async () => {
    if (!config.server) {
      setError('ArgoCD server is required');
      return;
    }

    // Check if Wails bindings are available
    if (!window.go || !window.go.main || !window.go.main.App) {
      setError('Wails bindings not available. Please wait for the app to fully load.');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const argoApps = await window.go.main.App.GetArgoApps(config);
      
      if (Array.isArray(argoApps)) {
        setApps(argoApps);
      } else {
        setError(`Expected array, got ${typeof argoApps}`);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      console.error('Failed to load apps:', error);
      
      // Check if this is an authentication error
      if (errorMessage.includes('authentication required') || errorMessage.includes('SAML redirect')) {
        setError(`${errorMessage}`);
      } else {
        setError(`Failed to load applications: ${errorMessage}`);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    loadApps();
  };

  // Load AWS profile on startup
  useEffect(() => {
    const initializeAWS = () => {
      if (window.go && window.go.main && window.go.main.App) {
        loadAWSProfile();
      } else {
        // Wait for Wails to be ready
        setTimeout(initializeAWS, 100);
      }
    };
    initializeAWS();
  }, []);

  // Auto-refresh every 30 seconds when enabled
  useEffect(() => {
    if (autoRefresh && config.server) {
      const interval = setInterval(loadApps, 30000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, config.server]);

  // Wait for Wails to be ready, then load apps when config changes
  useEffect(() => {
    const checkWailsAndLoad = () => {
      if (window.go && window.go.main && window.go.main.App && config.server && config.project) {
        loadApps();
      } else if (config.server && config.project) {
        // If config is set but Wails isn't ready, wait a bit and try again
        setTimeout(checkWailsAndLoad, 100);
      }
    };
    
    checkWailsAndLoad();
  }, [config.server, config.project]);

  const filteredApps = apps;
  
  const healthStats = apps.reduce((acc, app) => {
    if (app && app.Health) {
      const key = app.Health.toLowerCase();
      acc[key] = (acc[key] || 0) + 1;
    }
    return acc;
  }, {} as Record<string, number>);

  const syncStats = apps.reduce((acc, app) => {
    if (app && app.Sync) {
      const key = app.Sync.toLowerCase();
      acc[key] = (acc[key] || 0) + 1;
    }
    return acc;
  }, {} as Record<string, number>);

  const ArgocdInterface = () => (
    <>
      <header className="bg-slate-800 border-b border-slate-700 px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <ServerIcon className="w-8 h-8 text-blue-500" />
            <div>
              <h1 className="text-xl font-bold">Yak ArgoCD GUI</h1>
              <p className="text-sm text-slate-400">
                {config.server ? `Connected to ${config.server}` : 'Not connected'}
                {awsProfile && <span className="ml-2 px-2 py-1 bg-blue-600 rounded text-xs">AWS: {awsProfile}</span>}
              </p>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <label className="flex items-center space-x-2 text-sm">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="rounded border-slate-600 bg-slate-700 text-blue-600 focus:ring-blue-500"
              />
              <span>Auto-refresh</span>
            </label>

            <button
              onClick={loadAWSProfile}
              className="flex items-center space-x-2 bg-green-600 hover:bg-green-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
              title="Auto-detect ArgoCD server from AWS_PROFILE"
            >
              <span>Auto-detect</span>
            </button>

            <button
              onClick={() => setShowConfig(!showConfig)}
              className="flex items-center space-x-2 bg-slate-600 hover:bg-slate-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              <CogIcon className="w-4 h-4" />
              <span>Config</span>
            </button>

            <button
              onClick={handleRefresh}
              disabled={loading}
              className="flex items-center space-x-2 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 disabled:opacity-50 px-4 py-2 rounded-md text-sm font-medium transition-colors"
            >
              {loading ? (
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
              ) : (
                <ArrowPathIcon className="w-4 h-4" />
              )}
              <span>Refresh</span>
            </button>
          </div>
        </div>

        {showConfig && (
          <div className="mt-4 p-4 bg-slate-700 rounded-lg">
            <h3 className="text-lg font-medium mb-4">ArgoCD Configuration</h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-2">ArgoCD Server</label>
                <input
                  type="text"
                  value={config.server}
                  onChange={(e) => setConfig({ ...config, server: e.target.value })}
                  placeholder="argocd.example.com"
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">Project</label>
                <input
                  type="text"
                  value={config.project}
                  onChange={(e) => setConfig({ ...config, project: e.target.value })}
                  placeholder="main"
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>
          </div>
        )}
      </header>

      {error && (
        <div className="mx-6 mt-4 p-4 bg-red-900 border border-red-700 rounded-lg">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <XCircleIcon className="w-5 h-5 text-red-400" />
              <span className="text-red-100">{error}</span>
            </div>
            {(error.includes('authentication required') || error.includes('SAML redirect')) && (
              <button
                onClick={handleLogin}
                disabled={isLoggingIn}
                className="flex items-center space-x-2 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 disabled:opacity-50 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
              >
                {isLoggingIn ? (
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                ) : null}
                <span>{isLoggingIn ? 'Logging in...' : 'Login to ArgoCD'}</span>
              </button>
            )}
          </div>
        </div>
      )}

      <main className="p-6">

        {apps.length > 0 && (
          <div className="grid grid-cols-4 gap-4 mb-6">
            <div className="bg-slate-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-slate-400 mb-2">Total Applications</h3>
              <p className="text-2xl font-bold">{apps.length}</p>
            </div>
            <div className="bg-slate-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-slate-400 mb-2">Healthy</h3>
              <p className="text-2xl font-bold text-green-500">{healthStats.healthy || 0}</p>
            </div>
            <div className="bg-slate-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-slate-400 mb-2">Synced</h3>
              <p className="text-2xl font-bold text-green-500">{syncStats.synced || 0}</p>
            </div>
            <div className="bg-slate-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-slate-400 mb-2">Out of Sync</h3>
              <p className="text-2xl font-bold text-red-500">{syncStats.outofsync || 0}</p>
            </div>
          </div>
        )}

        {loading && apps.length === 0 ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
              <p className="text-slate-400">Loading applications...</p>
            </div>
          </div>
        ) : filteredApps.length === 0 ? (
          <div className="text-center py-12">
            <ServerIcon className="w-12 h-12 text-slate-500 mx-auto mb-4" />
            <p className="text-slate-400">
              {config.server ? 'No applications found' : 'Configure ArgoCD server to get started'}
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {filteredApps.map((app, index) => (
              <AppCard
                key={app?.AppName || `app-${index}`}
                app={app}
                config={config}
                onAction={handleRefresh}
              />
            ))}
          </div>
        )}
      </main>
    </>
  );

  return (
    <div className="min-h-screen bg-slate-900 text-white">
      {/* Tab Navigation */}
      <div className="bg-slate-800 border-b border-slate-700">
        <div className="px-6">
          <div className="flex space-x-8">
            <button
              onClick={() => setActiveTab('argocd')}
              className={`py-4 text-sm font-medium border-b-2 transition-colors ${
                activeTab === 'argocd'
                  ? 'border-blue-500 text-blue-400'
                  : 'border-transparent text-slate-400 hover:text-slate-300'
              }`}
            >
              ArgoCD Applications
            </button>
            <button
              onClick={() => setActiveTab('rollouts')}
              className={`py-4 text-sm font-medium border-b-2 transition-colors ${
                activeTab === 'rollouts'
                  ? 'border-purple-500 text-purple-400'
                  : 'border-transparent text-slate-400 hover:text-slate-300'
              }`}
            >
              Argo Rollouts
            </button>
            <button
              onClick={() => setActiveTab('secrets')}
              className={`py-4 text-sm font-medium border-b-2 transition-colors ${
                activeTab === 'secrets'
                  ? 'border-yellow-500 text-yellow-400'
                  : 'border-transparent text-slate-400 hover:text-slate-300'
              }`}
            >
              Secrets
            </button>
            <button
              onClick={() => setActiveTab('environment')}
              className={`py-4 text-sm font-medium border-b-2 transition-colors ${
                activeTab === 'environment'
                  ? 'border-green-500 text-green-400'
                  : 'border-transparent text-slate-400 hover:text-slate-300'
              }`}
            >
              Environment
            </button>
          </div>
        </div>
      </div>

      {/* Tab Content */}
      {activeTab === 'argocd' ? <ArgocdInterface /> : 
       activeTab === 'rollouts' ? <Rollouts /> : 
       activeTab === 'secrets' ? <Secrets /> :
       <EnvironmentConfig />}
    </div>
  );
};

export default App;