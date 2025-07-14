import React, { useState, useEffect } from 'react';
import { 
  ArrowPathIcon, 
  PlusIcon,
  PencilIcon,
  TrashIcon,
  EyeIcon,
  EyeSlashIcon,
  KeyIcon,
  LockClosedIcon,
  FolderIcon,
  ExclamationTriangleIcon,
  XCircleIcon
} from '@heroicons/react/24/outline';

// Types matching the Go backend
interface SecretListItem {
  path: string;
  version: number;
  owner: string;
  usage: string;
  source: string;
  createdAt: string;
  updatedAt: string;
}

interface SecretData {
  path: string;
  version: number;
  data: Record<string, string>;
  metadata: SecretMetadata;
}

interface SecretMetadata {
  owner: string;
  usage: string;
  source: string;
  createdAt: string;
  updatedAt: string;
  version: number;
  destroyed: boolean;
}

interface SecretConfig {
  platform: string;
  environment: string;
  team: string;
}

interface JWTClientConfig {
  platform: string;
  environment: string;
  team: string;
  path: string;
  owner: string;
  localName: string;
  targetService: string;
  secret: string;
}

interface JWTServerConfig {
  platform: string;
  environment: string;
  team: string;
  path: string;
  owner: string;
  localName: string;
  serviceName: string;
  clientName: string;
  clientSecret: string;
}


const SecretCard: React.FC<{ 
  secret: SecretListItem; 
  config: SecretConfig;
  onAction: () => void;
  onView: (secret: SecretListItem) => void;
  onEdit: (secret: SecretListItem) => void;
  onDelete: (secret: SecretListItem) => void;
  onNavigate: (path: string) => void;
}> = ({ secret, onView, onEdit, onDelete, onNavigate }) => {
  const formatDate = (dateStr: string) => {
    if (!dateStr) return 'Unknown';
    try {
      return new Date(dateStr).toLocaleDateString();
    } catch {
      return 'Unknown';
    }
  };

  const getSourceColor = (source: string) => {
    switch (source.toLowerCase()) {
      case 'manual': return 'bg-blue-600';
      case 'generated': return 'bg-green-600';
      case 'imported': return 'bg-purple-600';
      default: return 'bg-gray-600';
    }
  };

  const isFolder = secret?.path?.endsWith('/');
  const displayName = isFolder ? secret.path.slice(0, -1) : secret.path;

  const handleCardClick = () => {
    if (isFolder) {
      onNavigate(secret.path);
    }
  };

  return (
    <div 
      className={`bg-slate-800 rounded-lg p-6 border border-slate-700 hover:border-slate-600 transition-colors ${isFolder ? 'cursor-pointer' : ''}`}
      onClick={isFolder ? handleCardClick : undefined}
    >
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center space-x-3">
          {isFolder ? (
            <FolderIcon className="w-5 h-5 text-blue-500" />
          ) : (
            <KeyIcon className="w-5 h-5 text-yellow-500" />
          )}
          <div>
            <h3 className="text-lg font-semibold text-white font-mono">{displayName}</h3>
            <p className="text-sm text-slate-400">{isFolder ? 'Folder' : `v${secret?.version || 0}`}</p>
          </div>
        </div>
        {!isFolder && (
          <span className={`px-2 py-1 rounded-full text-xs font-medium text-white ${getSourceColor(secret?.source || '')}`}>
            {secret?.source || 'Unknown'}
          </span>
        )}
      </div>

      <div className="space-y-3 mb-4">
        <div className="flex items-center justify-between">
          <span className="text-sm text-slate-400">Owner:</span>
          <span className="text-sm text-slate-300">
            {secret?.owner && secret.owner !== 'Unknown' ? secret.owner : 'Click View for details'}
          </span>
        </div>
        
        <div className="flex items-center justify-between">
          <span className="text-sm text-slate-400">Usage:</span>
          <span className="text-sm text-slate-300 truncate max-w-48">
            {secret?.usage && secret.usage !== 'Unknown' ? secret.usage : 'Click View for details'}
          </span>
        </div>

        <div className="flex items-center justify-between">
          <span className="text-sm text-slate-400">Type:</span>
          <span className="text-sm text-slate-300">
            {secret?.path?.endsWith('/') ? 'Folder' : 'Secret'}
          </span>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-2">
        {isFolder ? (
          <button
            onClick={(e) => {
              e.stopPropagation();
              onNavigate(secret.path);
            }}
            className="col-span-3 flex items-center justify-center space-x-1 bg-blue-600 hover:bg-blue-700 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
          >
            <FolderIcon className="w-4 h-4" />
            <span>Open Folder</span>
          </button>
        ) : (
          <>
            <button
              onClick={(e) => {
                e.stopPropagation();
                onView(secret);
              }}
              className="flex items-center justify-center space-x-1 bg-blue-600 hover:bg-blue-700 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              <EyeIcon className="w-4 h-4" />
              <span>View</span>
            </button>

            <button
              onClick={(e) => {
                e.stopPropagation();
                onEdit(secret);
              }}
              className="flex items-center justify-center space-x-1 bg-green-600 hover:bg-green-700 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              <PencilIcon className="w-4 h-4" />
              <span>Edit</span>
            </button>

            <button
              onClick={(e) => {
                e.stopPropagation();
                onDelete(secret);
              }}
              className="flex items-center justify-center space-x-1 bg-red-600 hover:bg-red-700 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              <TrashIcon className="w-4 h-4" />
              <span>Delete</span>
            </button>
          </>
        )}
      </div>
    </div>
  );
};

const Secrets: React.FC = () => {
  const [secrets, setSecrets] = useState<SecretListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [config, setConfig] = useState<SecretConfig>({
    platform: 'dev',
    environment: '',
    team: ''
  });
  const [currentPath, setCurrentPath] = useState<string>('');
  const [error, setError] = useState<string | null>(null);
  const [showConfig, setShowConfig] = useState(false);
  const [selectedSecret, setSelectedSecret] = useState<SecretListItem | null>(null);
  const [secretData, setSecretData] = useState<SecretData | null>(null);
  const [showSecretDialog, setShowSecretDialog] = useState(false);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showJWTClientDialog, setShowJWTClientDialog] = useState(false);
  const [showJWTServerDialog, setShowJWTServerDialog] = useState(false);
  const [showValues, setShowValues] = useState(false);
  const [newSecretForm, setNewSecretForm] = useState({
    path: '',
    owner: '',
    usage: '',
    source: 'manual',
    data: {} as Record<string, string>
  });
  const [editSecretForm, setEditSecretForm] = useState({
    data: {} as Record<string, string>
  });
  const [jwtClientForm, setJwtClientForm] = useState<JWTClientConfig>({
    platform: 'dev',
    environment: '',
    team: '',
    path: '',
    owner: '',
    localName: '',
    targetService: '',
    secret: ''
  });
  const [jwtServerForm, setJwtServerForm] = useState<JWTServerConfig>({
    platform: 'dev',
    environment: '',
    team: '',
    path: '',
    owner: '',
    localName: '',
    serviceName: '',
    clientName: '',
    clientSecret: ''
  });
  const [breadcrumbs, setBreadcrumbs] = useState<string[]>([]);

  const handleNavigate = (path: string) => {
    setCurrentPath(path);
    // Build breadcrumbs from the current path
    const pathParts = path.split('/').filter(part => part !== '');
    setBreadcrumbs(pathParts);
  };

  const handleBreadcrumbClick = (index: number) => {
    const newPath = breadcrumbs.slice(0, index + 1).join('/') + '/';
    setCurrentPath(newPath);
    setBreadcrumbs(breadcrumbs.slice(0, index + 1));
  };

  const handleGoBack = () => {
    if (breadcrumbs.length > 0) {
      const newBreadcrumbs = breadcrumbs.slice(0, -1);
      const newPath = newBreadcrumbs.length > 0 ? newBreadcrumbs.join('/') + '/' : '';
      setCurrentPath(newPath);
      setBreadcrumbs(newBreadcrumbs);
    } else {
      setCurrentPath('');
      setBreadcrumbs([]);
    }
  };

  const loadSecrets = async () => {
    if (!config.platform) {
      setError('Platform is required');
      return;
    }

    if (!window.go || !window.go.main || !window.go.main.App) {
      setError('Wails bindings not available. Please wait for the app to fully load.');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const secretsData = await window.go.main.App.GetSecrets(config, currentPath);
      
      console.log('Secrets data received:', secretsData);
      console.log('Secrets data type:', typeof secretsData);
      console.log('Is array:', Array.isArray(secretsData));
      
      if (Array.isArray(secretsData)) {
        setSecrets(secretsData);
      } else {
        console.error('Received non-array data:', secretsData);
        setError(`Expected array, got ${typeof secretsData}`);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      console.error('Failed to load secrets:', error);
      setError(`Failed to load secrets: ${errorMessage}`);
    } finally {
      setLoading(false);
    }
  };

  const handleViewSecret = async (secret: SecretListItem) => {
    setSelectedSecret(secret);
    try {
      // Construct the full path by combining current path with secret path
      const fullPath = currentPath ? `${currentPath}${secret.path}` : secret.path;
      console.log('Getting secret with full path:', fullPath);
      const data = await window.go.main.App.GetSecretData(config, fullPath, secret.version);
      setSecretData(data);
      setShowSecretDialog(true);
    } catch (error) {
      setError(`Failed to load secret data: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  const handleCreateSecret = async () => {
    try {
      await window.go.main.App.CreateSecret(
        config,
        newSecretForm.path,
        newSecretForm.owner,
        newSecretForm.usage,
        newSecretForm.source,
        newSecretForm.data
      );
      setShowCreateDialog(false);
      setNewSecretForm({ path: '', owner: '', usage: '', source: 'manual', data: {} });
      loadSecrets();
    } catch (error) {
      setError(`Failed to create secret: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  const handleEditSecret = async (secret: SecretListItem) => {
    setSelectedSecret(secret);
    try {
      // Load current secret data for editing
      const fullPath = currentPath ? `${currentPath}${secret.path}` : secret.path;
      const data = await window.go.main.App.GetSecretData(config, fullPath, secret.version);
      setEditSecretForm({ data: data.data });
      setShowEditDialog(true);
    } catch (error) {
      setError(`Failed to load secret for editing: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  const handleUpdateSecret = async () => {
    if (!selectedSecret) return;
    
    try {
      const fullPath = currentPath ? `${currentPath}${selectedSecret.path}` : selectedSecret.path;
      await window.go.main.App.UpdateSecret(config, fullPath, editSecretForm.data);
      setShowEditDialog(false);
      setSelectedSecret(null);
      setEditSecretForm({ data: {} });
      loadSecrets();
    } catch (error) {
      setError(`Failed to update secret: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  const handleDeleteSecret = async () => {
    if (!selectedSecret) return;
    
    try {
      // Construct the full path by combining current path with secret path
      const fullPath = currentPath ? `${currentPath}${selectedSecret.path}` : selectedSecret.path;
      console.log('Deleting secret with full path:', fullPath);
      await window.go.main.App.DeleteSecret(config, fullPath, selectedSecret.version);
      setShowDeleteDialog(false);
      setSelectedSecret(null);
      loadSecrets();
    } catch (error) {
      setError(`Failed to delete secret: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  const addDataField = () => {
    const key = `key${Object.keys(newSecretForm.data).length + 1}`;
    setNewSecretForm(prev => ({
      ...prev,
      data: { ...prev.data, [key]: '' }
    }));
  };

  const updateDataField = (oldKey: string, newKey: string, value: string) => {
    setNewSecretForm(prev => {
      const newData = { ...prev.data };
      if (oldKey !== newKey && newData[oldKey] !== undefined) {
        delete newData[oldKey];
      }
      newData[newKey] = value;
      return { ...prev, data: newData };
    });
  };

  const removeDataField = (key: string) => {
    setNewSecretForm(prev => {
      const newData = { ...prev.data };
      delete newData[key];
      return { ...prev, data: newData };
    });
  };

  const addEditDataField = () => {
    const key = `key${Object.keys(editSecretForm.data).length + 1}`;
    setEditSecretForm(prev => ({
      ...prev,
      data: { ...prev.data, [key]: '' }
    }));
  };

  const updateEditDataField = (oldKey: string, newKey: string, value: string) => {
    setEditSecretForm(prev => {
      const newData = { ...prev.data };
      if (oldKey !== newKey && newData[oldKey] !== undefined) {
        delete newData[oldKey];
      }
      newData[newKey] = value;
      return { ...prev, data: newData };
    });
  };

  const removeEditDataField = (key: string) => {
    setEditSecretForm(prev => {
      const newData = { ...prev.data };
      delete newData[key];
      return { ...prev, data: newData };
    });
  };

  const handleCreateJWTClient = async () => {
    try {
      await window.go.main.App.CreateJWTClient(jwtClientForm);
      setShowJWTClientDialog(false);
      setJwtClientForm({
        platform: 'dev',
        environment: '',
        team: '',
        path: '',
        owner: '',
        localName: '',
        targetService: '',
        secret: ''
      });
      loadSecrets();
    } catch (error) {
      setError(`Failed to create JWT client: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  const handleCreateJWTServer = async () => {
    try {
      await window.go.main.App.CreateJWTServer(jwtServerForm);
      setShowJWTServerDialog(false);
      setJwtServerForm({
        platform: 'dev',
        environment: '',
        team: '',
        path: '',
        owner: '',
        localName: '',
        serviceName: '',
        clientName: '',
        clientSecret: ''
      });
      loadSecrets();
    } catch (error) {
      setError(`Failed to create JWT server: ${error instanceof Error ? error.message : String(error)}`);
    }
  };

  useEffect(() => {
    loadSecrets();
  }, [config, currentPath]);

  const platforms = ['dev', 'staging', 'production'];

  return (
    <div className="min-h-screen bg-slate-900 text-white">
      <header className="bg-slate-800 border-b border-slate-700 px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <LockClosedIcon className="w-8 h-8 text-yellow-500" />
            <div>
              <h1 className="text-xl font-bold">Yak Secrets GUI</h1>
              <p className="text-sm text-slate-400">
                Platform: {config.platform}
                {currentPath && <span className="ml-2 px-2 py-1 bg-yellow-600 rounded text-xs">Path: {currentPath}</span>}
              </p>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <button
              onClick={() => setShowCreateDialog(true)}
              className="flex items-center space-x-2 bg-green-600 hover:bg-green-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              <PlusIcon className="w-4 h-4" />
              <span>Create</span>
            </button>

            <button
              onClick={() => setShowJWTClientDialog(true)}
              className="flex items-center space-x-2 bg-blue-600 hover:bg-blue-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              <KeyIcon className="w-4 h-4" />
              <span>JWT Client</span>
            </button>

            <button
              onClick={() => setShowJWTServerDialog(true)}
              className="flex items-center space-x-2 bg-purple-600 hover:bg-purple-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              <LockClosedIcon className="w-4 h-4" />
              <span>JWT Server</span>
            </button>

            <button
              onClick={() => setShowConfig(!showConfig)}
              className="flex items-center space-x-2 bg-slate-600 hover:bg-slate-700 px-3 py-2 rounded-md text-sm font-medium transition-colors"
            >
              <span>Config</span>
            </button>

            <button
              onClick={loadSecrets}
              disabled={loading}
              className="flex items-center space-x-2 bg-yellow-600 hover:bg-yellow-700 disabled:bg-yellow-800 disabled:opacity-50 px-4 py-2 rounded-md text-sm font-medium transition-colors"
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
            <h3 className="text-lg font-medium mb-4">Secret Configuration</h3>
            <div className="grid grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium mb-2">Platform</label>
                <select
                  value={config.platform}
                  onChange={(e) => setConfig({ ...config, platform: e.target.value })}
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-yellow-500"
                >
                  {platforms.map(platform => (
                    <option key={platform} value={platform}>{platform}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">Environment (optional)</label>
                <input
                  type="text"
                  value={config.environment}
                  onChange={(e) => setConfig({ ...config, environment: e.target.value })}
                  placeholder="feature, hotfix, etc."
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-yellow-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">Team (optional)</label>
                <input
                  type="text"
                  value={config.team}
                  onChange={(e) => setConfig({ ...config, team: e.target.value })}
                  placeholder="backend, frontend, etc."
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-yellow-500"
                />
              </div>
            </div>
            <div className="mt-4">
              <label className="block text-sm font-medium mb-2">Secret Path (optional)</label>
              <input
                type="text"
                value={currentPath}
                onChange={(e) => setCurrentPath(e.target.value)}
                placeholder="myapp/, myapp/database, etc."
                className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-yellow-500"
              />
            </div>
          </div>
        )}
      </header>

      {error && (
        <div className="mx-6 mt-4 p-4 bg-red-900 border border-red-700 rounded-lg">
          <div className="flex items-center space-x-2">
            <XCircleIcon className="w-5 h-5 text-red-400" />
            <span className="text-red-100">{error}</span>
          </div>
        </div>
      )}

      {/* Breadcrumb Navigation */}
      {(breadcrumbs.length > 0 || currentPath) && (
        <div className="mx-6 mt-4 p-3 bg-slate-800 rounded-lg border border-slate-700">
          <div className="flex items-center space-x-2">
            <button
              onClick={handleGoBack}
              className="flex items-center space-x-1 bg-slate-600 hover:bg-slate-700 px-2 py-1 rounded text-sm transition-colors"
            >
              <span>←</span>
              <span>Back</span>
            </button>
            
            <span className="text-slate-400">/</span>
            
            <button
              onClick={() => {
                setCurrentPath('');
                setBreadcrumbs([]);
              }}
              className="text-yellow-400 hover:text-yellow-300 text-sm transition-colors"
            >
              root
            </button>
            
            {breadcrumbs.map((crumb, index) => (
              <React.Fragment key={index}>
                <span className="text-slate-400">/</span>
                <button
                  onClick={() => handleBreadcrumbClick(index)}
                  className="text-yellow-400 hover:text-yellow-300 text-sm transition-colors"
                >
                  {crumb}
                </button>
              </React.Fragment>
            ))}
          </div>
        </div>
      )}

      <main className="p-6">
        {secrets.length > 0 && (
          <div className="grid grid-cols-4 gap-4 mb-6">
            <div className="bg-slate-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-slate-400 mb-2">Total Secrets</h3>
              <p className="text-2xl font-bold">{secrets.length}</p>
            </div>
            <div className="bg-slate-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-slate-400 mb-2">Folders</h3>
              <p className="text-2xl font-bold text-blue-500">
                {secrets.filter(s => s.path?.endsWith('/')).length}
              </p>
            </div>
            <div className="bg-slate-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-slate-400 mb-2">Secrets</h3>
              <p className="text-2xl font-bold text-green-500">
                {secrets.filter(s => !s.path?.endsWith('/')).length}
              </p>
            </div>
            <div className="bg-slate-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-slate-400 mb-2">Paths</h3>
              <p className="text-2xl font-bold text-purple-500">
                {secrets.length}
              </p>
            </div>
          </div>
        )}

        {loading && secrets.length === 0 ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="w-8 h-8 border-2 border-yellow-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
              <p className="text-slate-400">Loading secrets...</p>
            </div>
          </div>
        ) : secrets.length === 0 ? (
          <div className="text-center py-12">
            <LockClosedIcon className="w-12 h-12 text-slate-500 mx-auto mb-4" />
            <p className="text-slate-400">No secrets found</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {secrets.map((secret, index) => (
              <SecretCard
                key={secret?.path || `secret-${index}`}
                secret={secret}
                config={config}
                onAction={loadSecrets}
                onView={handleViewSecret}
                onEdit={handleEditSecret}
                onDelete={(secret) => { setSelectedSecret(secret); setShowDeleteDialog(true); }}
                onNavigate={handleNavigate}
              />
            ))}
          </div>
        )}
      </main>

      {/* View Secret Dialog */}
      {showSecretDialog && secretData && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-slate-800 rounded-lg p-6 w-[600px] max-h-[80vh] overflow-y-auto border border-slate-700">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-white">Secret: {secretData.path}</h3>
              <button
                onClick={() => setShowValues(!showValues)}
                className="flex items-center space-x-2 bg-blue-600 hover:bg-blue-700 px-3 py-2 rounded-md text-sm"
              >
                {showValues ? <EyeSlashIcon className="w-4 h-4" /> : <EyeIcon className="w-4 h-4" />}
                <span>{showValues ? 'Hide' : 'Show'}</span>
              </button>
            </div>
            
            <div className="space-y-4">
              <div>
                <h4 className="text-sm font-medium text-slate-400 mb-2">Metadata</h4>
                <div className="text-xs space-y-1 text-slate-300">
                  <p>Owner: {secretData.metadata.owner}</p>
                  <p>Usage: {secretData.metadata.usage}</p>
                  <p>Source: {secretData.metadata.source}</p>
                  <p>Version: {secretData.metadata.version}</p>
                </div>
              </div>

              <div>
                <h4 className="text-sm font-medium text-slate-400 mb-2">Data</h4>
                <div className="space-y-2">
                  {Object.entries(secretData.data).map(([key, value]) => (
                    <div key={key} className="flex justify-between">
                      <span className="text-sm text-slate-300 font-mono">{key}:</span>
                      <span className="text-sm text-white font-mono">
                        {showValues ? value : '••••••••'}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="flex justify-end mt-6">
              <button
                onClick={() => setShowSecretDialog(false)}
                className="px-4 py-2 bg-slate-600 hover:bg-slate-700 text-white rounded-md text-sm font-medium transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Create Secret Dialog */}
      {showCreateDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-slate-800 rounded-lg p-6 w-[600px] max-h-[80vh] overflow-y-auto border border-slate-700">
            <h3 className="text-lg font-semibold text-white mb-4">Create New Secret</h3>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-2">Path</label>
                <input
                  type="text"
                  value={newSecretForm.path}
                  onChange={(e) => setNewSecretForm({ ...newSecretForm, path: e.target.value })}
                  placeholder="myapp/database"
                  className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-yellow-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-400 mb-2">Owner</label>
                <input
                  type="text"
                  value={newSecretForm.owner}
                  onChange={(e) => setNewSecretForm({ ...newSecretForm, owner: e.target.value })}
                  placeholder="team-backend"
                  className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-yellow-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-400 mb-2">Usage</label>
                <input
                  type="text"
                  value={newSecretForm.usage}
                  onChange={(e) => setNewSecretForm({ ...newSecretForm, usage: e.target.value })}
                  placeholder="Database credentials"
                  className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-yellow-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-400 mb-2">Source</label>
                <select
                  value={newSecretForm.source}
                  onChange={(e) => setNewSecretForm({ ...newSecretForm, source: e.target.value })}
                  className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-yellow-500"
                >
                  <option value="manual">Manual</option>
                  <option value="generated">Generated</option>
                  <option value="imported">Imported</option>
                </select>
              </div>

              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-sm font-medium text-slate-400">Data</label>
                  <button
                    onClick={addDataField}
                    className="flex items-center space-x-1 bg-green-600 hover:bg-green-700 px-2 py-1 rounded text-xs"
                  >
                    <PlusIcon className="w-3 h-3" />
                    <span>Add</span>
                  </button>
                </div>
                <div className="space-y-2">
                  {Object.entries(newSecretForm.data).map(([key, value]) => (
                    <div key={key} className="flex space-x-2">
                      <input
                        type="text"
                        value={key}
                        onChange={(e) => updateDataField(key, e.target.value, value)}
                        placeholder="key"
                        className="flex-1 px-2 py-1 bg-slate-700 border border-slate-600 rounded text-white text-sm focus:outline-none focus:ring-1 focus:ring-yellow-500"
                      />
                      <input
                        type="text"
                        value={value}
                        onChange={(e) => updateDataField(key, key, e.target.value)}
                        placeholder="value"
                        className="flex-1 px-2 py-1 bg-slate-700 border border-slate-600 rounded text-white text-sm focus:outline-none focus:ring-1 focus:ring-yellow-500"
                      />
                      <button
                        onClick={() => removeDataField(key)}
                        className="px-2 py-1 bg-red-600 hover:bg-red-700 rounded text-xs"
                      >
                        <TrashIcon className="w-3 h-3" />
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="flex space-x-3 mt-6">
              <button
                onClick={handleCreateSecret}
                disabled={!newSecretForm.path || !newSecretForm.owner || !newSecretForm.usage}
                className="flex-1 bg-green-600 hover:bg-green-700 disabled:bg-green-800 disabled:opacity-50 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors"
              >
                Create
              </button>
              
              <button
                onClick={() => {
                  setShowCreateDialog(false);
                  setNewSecretForm({ path: '', owner: '', usage: '', source: 'manual', data: {} });
                }}
                className="px-4 py-2 bg-slate-600 hover:bg-slate-700 text-white rounded-md text-sm font-medium transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Secret Dialog */}
      {showEditDialog && selectedSecret && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-slate-800 rounded-lg p-6 w-[600px] max-h-[80vh] overflow-y-auto border border-slate-700">
            <h3 className="text-lg font-semibold text-white mb-4">Edit Secret: {selectedSecret.path}</h3>
            
            <div className="space-y-4">
              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="block text-sm font-medium text-slate-400">Data</label>
                  <button
                    onClick={addEditDataField}
                    className="flex items-center space-x-1 bg-green-600 hover:bg-green-700 px-2 py-1 rounded text-xs"
                  >
                    <PlusIcon className="w-3 h-3" />
                    <span>Add</span>
                  </button>
                </div>
                <div className="space-y-2">
                  {Object.entries(editSecretForm.data).map(([key, value]) => (
                    <div key={key} className="flex space-x-2">
                      <input
                        type="text"
                        value={key}
                        onChange={(e) => updateEditDataField(key, e.target.value, value)}
                        placeholder="key"
                        className="flex-1 px-2 py-1 bg-slate-700 border border-slate-600 rounded text-white text-sm focus:outline-none focus:ring-1 focus:ring-yellow-500"
                      />
                      <input
                        type="text"
                        value={value}
                        onChange={(e) => updateEditDataField(key, key, e.target.value)}
                        placeholder="value"
                        className="flex-1 px-2 py-1 bg-slate-700 border border-slate-600 rounded text-white text-sm focus:outline-none focus:ring-1 focus:ring-yellow-500"
                      />
                      <button
                        onClick={() => removeEditDataField(key)}
                        className="px-2 py-1 bg-red-600 hover:bg-red-700 rounded text-xs"
                      >
                        <TrashIcon className="w-3 h-3" />
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="flex space-x-3 mt-6">
              <button
                onClick={handleUpdateSecret}
                disabled={Object.keys(editSecretForm.data).length === 0}
                className="flex-1 bg-green-600 hover:bg-green-700 disabled:bg-green-800 disabled:opacity-50 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors"
              >
                Update
              </button>
              
              <button
                onClick={() => {
                  setShowEditDialog(false);
                  setSelectedSecret(null);
                  setEditSecretForm({ data: {} });
                }}
                className="px-4 py-2 bg-slate-600 hover:bg-slate-700 text-white rounded-md text-sm font-medium transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {showDeleteDialog && selectedSecret && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-slate-800 rounded-lg p-6 w-[500px] border border-slate-700">
            <div className="flex items-center space-x-3 mb-4">
              <ExclamationTriangleIcon className="w-6 h-6 text-red-500" />
              <h3 className="text-lg font-semibold text-white">Delete Secret</h3>
            </div>
            
            <p className="text-slate-300 mb-6">
              Are you sure you want to delete secret <span className="font-mono text-white">{selectedSecret.path}</span> version {selectedSecret.version}?
              This action cannot be undone.
            </p>

            <div className="flex space-x-3">
              <button
                onClick={handleDeleteSecret}
                className="flex-1 bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors"
              >
                Delete
              </button>
              
              <button
                onClick={() => {
                  setShowDeleteDialog(false);
                  setSelectedSecret(null);
                }}
                className="px-4 py-2 bg-slate-600 hover:bg-slate-700 text-white rounded-md text-sm font-medium transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* JWT Client Dialog */}
      {showJWTClientDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-slate-800 rounded-lg p-6 w-[700px] max-h-[80vh] overflow-y-auto border border-slate-700">
            <h3 className="text-lg font-semibold text-white mb-4">Create JWT Client Secret</h3>
            
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Platform</label>
                  <select
                    value={jwtClientForm.platform}
                    onChange={(e) => setJwtClientForm({ ...jwtClientForm, platform: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="dev">dev</option>
                    <option value="staging">staging</option>
                    <option value="production">production</option>
                  </select>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Environment</label>
                  <input
                    type="text"
                    value={jwtClientForm.environment}
                    onChange={(e) => setJwtClientForm({ ...jwtClientForm, environment: e.target.value })}
                    placeholder="feature, hotfix, etc."
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-400 mb-2">Path *</label>
                <input
                  type="text"
                  value={jwtClientForm.path}
                  onChange={(e) => setJwtClientForm({ ...jwtClientForm, path: e.target.value })}
                  placeholder="personal-assistant/jwt"
                  className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Owner *</label>
                  <input
                    type="text"
                    value={jwtClientForm.owner}
                    onChange={(e) => setJwtClientForm({ ...jwtClientForm, owner: e.target.value })}
                    placeholder="team-backend"
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Team</label>
                  <input
                    type="text"
                    value={jwtClientForm.team}
                    onChange={(e) => setJwtClientForm({ ...jwtClientForm, team: e.target.value })}
                    placeholder="backend, frontend, etc."
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Local Name *</label>
                  <input
                    type="text"
                    value={jwtClientForm.localName}
                    onChange={(e) => setJwtClientForm({ ...jwtClientForm, localName: e.target.value })}
                    placeholder="personal-assistant"
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Target Service *</label>
                  <input
                    type="text"
                    value={jwtClientForm.targetService}
                    onChange={(e) => setJwtClientForm({ ...jwtClientForm, targetService: e.target.value })}
                    placeholder="organization_admin"
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-400 mb-2">Secret *</label>
                <input
                  type="password"
                  value={jwtClientForm.secret}
                  onChange={(e) => setJwtClientForm({ ...jwtClientForm, secret: e.target.value })}
                  placeholder="HMAC-SHA256 Secret"
                  className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>

            <div className="flex space-x-3 mt-6">
              <button
                onClick={handleCreateJWTClient}
                disabled={!jwtClientForm.path || !jwtClientForm.owner || !jwtClientForm.localName || !jwtClientForm.targetService || !jwtClientForm.secret}
                className="flex-1 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 disabled:opacity-50 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors"
              >
                Create JWT Client
              </button>
              
              <button
                onClick={() => {
                  setShowJWTClientDialog(false);
                  setJwtClientForm({
                    platform: 'dev',
                    environment: '',
                    team: '',
                    path: '',
                    owner: '',
                    localName: '',
                    targetService: '',
                    secret: ''
                  });
                }}
                className="px-4 py-2 bg-slate-600 hover:bg-slate-700 text-white rounded-md text-sm font-medium transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* JWT Server Dialog */}
      {showJWTServerDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-slate-800 rounded-lg p-6 w-[700px] max-h-[80vh] overflow-y-auto border border-slate-700">
            <h3 className="text-lg font-semibold text-white mb-4">Create JWT Server Secret</h3>
            
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Platform</label>
                  <select
                    value={jwtServerForm.platform}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, platform: e.target.value })}
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
                  >
                    <option value="dev">dev</option>
                    <option value="staging">staging</option>
                    <option value="production">production</option>
                  </select>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Environment</label>
                  <input
                    type="text"
                    value={jwtServerForm.environment}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, environment: e.target.value })}
                    placeholder="feature, hotfix, etc."
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-400 mb-2">Path *</label>
                <input
                  type="text"
                  value={jwtServerForm.path}
                  onChange={(e) => setJwtServerForm({ ...jwtServerForm, path: e.target.value })}
                  placeholder="doctolib/secrets"
                  className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Owner *</label>
                  <input
                    type="text"
                    value={jwtServerForm.owner}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, owner: e.target.value })}
                    placeholder="team-backend"
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Team</label>
                  <input
                    type="text"
                    value={jwtServerForm.team}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, team: e.target.value })}
                    placeholder="backend, frontend, etc."
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Local Name *</label>
                  <input
                    type="text"
                    value={jwtServerForm.localName}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, localName: e.target.value })}
                    placeholder="monolith"
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Service Name *</label>
                  <input
                    type="text"
                    value={jwtServerForm.serviceName}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, serviceName: e.target.value })}
                    placeholder="organization_admin"
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Client Name *</label>
                  <input
                    type="text"
                    value={jwtServerForm.clientName}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, clientName: e.target.value })}
                    placeholder="personal-assistant"
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-400 mb-2">Client Secret *</label>
                  <input
                    type="password"
                    value={jwtServerForm.clientSecret}
                    onChange={(e) => setJwtServerForm({ ...jwtServerForm, clientSecret: e.target.value })}
                    placeholder="HMAC-SHA256 Secret"
                    className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
                  />
                </div>
              </div>
            </div>

            <div className="flex space-x-3 mt-6">
              <button
                onClick={handleCreateJWTServer}
                disabled={!jwtServerForm.path || !jwtServerForm.owner || !jwtServerForm.localName || !jwtServerForm.serviceName || !jwtServerForm.clientName || !jwtServerForm.clientSecret}
                className="flex-1 bg-purple-600 hover:bg-purple-700 disabled:bg-purple-800 disabled:opacity-50 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors"
              >
                Create JWT Server
              </button>
              
              <button
                onClick={() => {
                  setShowJWTServerDialog(false);
                  setJwtServerForm({
                    platform: 'dev',
                    environment: '',
                    team: '',
                    path: '',
                    owner: '',
                    localName: '',
                    serviceName: '',
                    clientName: '',
                    clientSecret: ''
                  });
                }}
                className="px-4 py-2 bg-slate-600 hover:bg-slate-700 text-white rounded-md text-sm font-medium transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Secrets;