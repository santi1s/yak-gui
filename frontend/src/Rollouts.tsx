import React, { useState, useEffect } from 'react';
import { 
  ArrowPathIcon, 
  PauseIcon, 
  StopIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  ServerIcon,
  CogIcon,
  ArrowUpIcon,
  ChevronDoubleUpIcon,
  PhotoIcon,
  EyeIcon,
  DocumentTextIcon,
  ClipboardDocumentIcon
} from '@heroicons/react/24/outline';

// Types matching the Go backend
interface RolloutListItem {
  name: string;
  namespace: string;
  status: string;
  replicas: string;
  age: string;
  strategy: string;
  revision: string;
  images: Record<string, string>;
}

interface RolloutStatus {
  name: string;
  namespace: string;
  status: string;
  replicas: string;
  updated: string;
  ready: string;
  available: string;
  strategy: string;
  currentStep: string;
  revision: string;
  message: string;
  analysis: string;
  images: Record<string, string>;
}

interface KubernetesConfig {
  server: string;
  namespace: string;
}

const StatusBadge: React.FC<{ status: string, type: 'status' | 'strategy' }> = ({ status, type }) => {
  const getStatusClass = () => {
    if (type === 'status') {
      switch (status.toLowerCase()) {
        case 'healthy': return 'bg-green-600 text-white';
        case 'progressing': return 'bg-blue-600 text-white';
        case 'degraded': return 'bg-red-600 text-white';
        case 'paused': return 'bg-yellow-600 text-white';
        case 'error': return 'bg-red-800 text-white';
        default: return 'bg-gray-600 text-white';
      }
    } else if (type === 'strategy') {
      switch (status.toLowerCase()) {
        case 'canary': return 'bg-purple-600 text-white';
        case 'bluegreen': return 'bg-teal-600 text-white';
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

// Detailed Rollout Modal Component
const RolloutDetailModal: React.FC<{ 
  rollout: RolloutListItem; 
  config: KubernetesConfig;
  isOpen: boolean;
  onClose: () => void;
}> = ({ rollout, config, isOpen, onClose }) => {
  const [detailedStatus, setDetailedStatus] = useState<RolloutStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadDetailedStatus = async () => {
    if (!isOpen || !rollout?.name) return;
    
    setLoading(true);
    setError(null);
    try {
      const status = await window.go.main.App.GetRolloutStatus(config, rollout.name);
      setDetailedStatus(status);
    } catch (error) {
      setError(`Failed to load detailed status: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      loadDetailedStatus();
    }
  }, [isOpen, rollout?.name]);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-slate-800 rounded-lg w-full max-w-4xl max-h-[90vh] overflow-y-auto border border-slate-700">
        <div className="sticky top-0 bg-slate-800 border-b border-slate-700 px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <DocumentTextIcon className="w-6 h-6 text-purple-400" />
              <div>
                <h2 className="text-xl font-bold text-white">{rollout?.name}</h2>
                <p className="text-sm text-slate-400">{rollout?.namespace}</p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="text-slate-400 hover:text-white"
            >
              <XCircleIcon className="w-6 h-6" />
            </button>
          </div>
        </div>

        <div className="p-6">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <div className="w-8 h-8 border-2 border-purple-500 border-t-transparent rounded-full animate-spin"></div>
            </div>
          ) : error ? (
            <div className="p-4 bg-red-900 border border-red-700 rounded-lg">
              <div className="flex items-center space-x-2">
                <XCircleIcon className="w-5 h-5 text-red-400" />
                <span className="text-red-100">{error}</span>
              </div>
            </div>
          ) : (
            <div className="space-y-6">
              {/* Basic Information */}
              <div className="bg-slate-700 rounded-lg p-4">
                <h3 className="text-lg font-semibold text-white mb-4">Basic Information</h3>
                <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                  <div>
                    <span className="text-sm text-slate-400">Status:</span>
                    <div className="mt-1">
                      <StatusBadge status={detailedStatus?.status || rollout?.status || 'Unknown'} type="status" />
                    </div>
                  </div>
                  <div>
                    <span className="text-sm text-slate-400">Strategy:</span>
                    <div className="mt-1">
                      <StatusBadge status={detailedStatus?.strategy || rollout?.strategy || 'Unknown'} type="strategy" />
                    </div>
                  </div>
                  <div>
                    <span className="text-sm text-slate-400">Replicas:</span>
                    <p className="text-white font-mono">{detailedStatus?.replicas || rollout?.replicas || 'N/A'}</p>
                  </div>
                  <div>
                    <span className="text-sm text-slate-400">Current Step:</span>
                    <p className="text-white font-mono">{detailedStatus?.currentStep || 'N/A'}</p>
                  </div>
                  <div>
                    <span className="text-sm text-slate-400">Revision:</span>
                    <p className="text-white font-mono">{detailedStatus?.revision || rollout?.revision || 'N/A'}</p>
                  </div>
                  <div>
                    <span className="text-sm text-slate-400">Age:</span>
                    <p className="text-white">{rollout?.age || 'N/A'}</p>
                  </div>
                </div>
              </div>

              {/* Detailed Status Information */}
              {detailedStatus && (
                <div className="bg-slate-700 rounded-lg p-4">
                  <h3 className="text-lg font-semibold text-white mb-4">Detailed Status</h3>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div>
                      <span className="text-sm text-slate-400">Updated:</span>
                      <p className="text-white font-mono">{detailedStatus.updated || 'N/A'}</p>
                    </div>
                    <div>
                      <span className="text-sm text-slate-400">Ready:</span>
                      <p className="text-white font-mono">{detailedStatus.ready || 'N/A'}</p>
                    </div>
                    <div>
                      <span className="text-sm text-slate-400">Available:</span>
                      <p className="text-white font-mono">{detailedStatus.available || 'N/A'}</p>
                    </div>
                    <div>
                      <span className="text-sm text-slate-400">Analysis:</span>
                      <p className="text-white font-mono">{detailedStatus.analysis || 'N/A'}</p>
                    </div>
                  </div>
                  {detailedStatus.message && (
                    <div className="mt-4">
                      <span className="text-sm text-slate-400">Message:</span>
                      <p className="text-white bg-slate-800 p-3 rounded mt-1 font-mono text-sm">{detailedStatus.message}</p>
                    </div>
                  )}
                </div>
              )}

              {/* Container Images - Full Details */}
              {((detailedStatus?.images && Object.keys(detailedStatus.images).length > 0) || 
                (rollout?.images && Object.keys(rollout.images).length > 0)) && (
                <div className="bg-slate-700 rounded-lg p-4">
                  <h3 className="text-lg font-semibold text-white mb-4 flex items-center space-x-2">
                    <PhotoIcon className="w-5 h-5 text-purple-400" />
                    <span>Container Images</span>
                  </h3>
                  <div className="space-y-4">
                    {Object.entries(detailedStatus?.images || rollout?.images || {}).map(([container, image]) => {
                      // Split by registry/path and image:tag
                      const imageParts = image.split('/');
                      const imageNameWithTag = imageParts[imageParts.length - 1];
                      const registry = imageParts.length > 1 ? imageParts.slice(0, -1).join('/') : '';
                      
                      // Handle both tag and digest formats
                      let imageName, version;
                      if (imageNameWithTag.includes('@sha256:')) {
                        [imageName, version] = imageNameWithTag.split('@');
                      } else if (imageNameWithTag.includes(':')) {
                        [imageName, version] = imageNameWithTag.split(':');
                      } else {
                        imageName = imageNameWithTag;
                        version = 'latest';
                      }
                      
                      return (
                        <div key={container} className="bg-slate-800 rounded-lg p-4">
                          <div className="flex items-center justify-between mb-3">
                            <h4 className="text-md font-medium text-white">{container}</h4>
                            <button
                              onClick={() => copyToClipboard(image)}
                              className="text-slate-400 hover:text-white"
                              title="Copy full image reference"
                            >
                              <ClipboardDocumentIcon className="w-4 h-4" />
                            </button>
                          </div>
                          
                          <div className="space-y-2">
                            <div>
                              <span className="text-xs text-slate-400">Registry:</span>
                              <p className="text-white font-mono text-sm break-all">{registry || 'docker.io'}</p>
                            </div>
                            <div>
                              <span className="text-xs text-slate-400">Image:</span>
                              <p className="text-white font-mono text-sm">{imageName}</p>
                            </div>
                            <div>
                              <span className="text-xs text-slate-400">Version:</span>
                              <div className="flex items-center space-x-2">
                                <span 
                                  className={`px-2 py-1 rounded font-mono text-sm break-all ${
                                    version && version.startsWith('sha256:') 
                                      ? 'text-orange-300 bg-orange-900' 
                                      : 'text-purple-300 bg-purple-900'
                                  }`}
                                >
                                  {version}
                                </span>
                                {version && version.startsWith('sha256:') && (
                                  <button
                                    onClick={() => copyToClipboard(version)}
                                    className="text-slate-400 hover:text-white"
                                    title="Copy SHA256 digest"
                                  >
                                    <ClipboardDocumentIcon className="w-4 h-4" />
                                  </button>
                                )}
                              </div>
                            </div>
                            <div>
                              <span className="text-xs text-slate-400">Full Reference:</span>
                              <div className="flex items-center space-x-2">
                                <p className="text-white font-mono text-sm bg-slate-900 p-2 rounded flex-1 break-all">{image}</p>
                                <button
                                  onClick={() => copyToClipboard(image)}
                                  className="text-slate-400 hover:text-white flex-shrink-0"
                                  title="Copy full image reference"
                                >
                                  <ClipboardDocumentIcon className="w-4 h-4" />
                                </button>
                              </div>
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

const RolloutCard: React.FC<{ 
  rollout: RolloutListItem; 
  config: KubernetesConfig;
  onAction: () => void;
}> = ({ rollout, config, onAction }) => {
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [showImageDialog, setShowImageDialog] = useState(false);
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [newImage, setNewImage] = useState('');
  const [containerName, setContainerName] = useState('');

  const handlePromote = async (full: boolean = false) => {
    setActionLoading(full ? 'promote-full' : 'promote');
    try {
      await window.go.main.App.PromoteRollout(config, rollout?.name || '', full);
      onAction();
    } catch (error) {
      console.error('Failed to promote rollout:', error);
    } finally {
      setActionLoading(null);
    }
  };

  const handlePause = async () => {
    setActionLoading('pause');
    try {
      await window.go.main.App.PauseRollout(config, rollout?.name || '');
      onAction();
    } catch (error) {
      console.error('Failed to pause rollout:', error);
    } finally {
      setActionLoading(null);
    }
  };

  const handleAbort = async () => {
    setActionLoading('abort');
    try {
      await window.go.main.App.AbortRollout(config, rollout?.name || '');
      onAction();
    } catch (error) {
      console.error('Failed to abort rollout:', error);
    } finally {
      setActionLoading(null);
    }
  };

  const handleRestart = async () => {
    setActionLoading('restart');
    try {
      await window.go.main.App.RestartRollout(config, rollout?.name || '');
      onAction();
    } catch (error) {
      console.error('Failed to restart rollout:', error);
    } finally {
      setActionLoading(null);
    }
  };

  const handleSetImage = async () => {
    if (!newImage) return;
    
    setActionLoading('set-image');
    try {
      await window.go.main.App.SetRolloutImage(config, rollout?.name || '', newImage, containerName);
      setShowImageDialog(false);
      setNewImage('');
      setContainerName('');
      onAction();
    } catch (error) {
      console.error('Failed to set image:', error);
    } finally {
      setActionLoading(null);
    }
  };

  const getStatusIcon = () => {
    const status = (rollout?.status || '').toLowerCase();
    switch (status) {
      case 'healthy':
        return <CheckCircleIcon className="w-5 h-5 text-green-500" />;
      case 'degraded':
      case 'error':
        return <XCircleIcon className="w-5 h-5 text-red-500" />;
      case 'progressing':
        return <ClockIcon className="w-5 h-5 text-blue-500" />;
      case 'paused':
        return <PauseIcon className="w-5 h-5 text-yellow-500" />;
      default:
        return <ExclamationTriangleIcon className="w-5 h-5 text-gray-500" />;
    }
  };

  return (
    <>
      <div className="bg-slate-800 rounded-lg p-6 border border-slate-700 hover:border-slate-600 transition-colors min-w-0">
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center space-x-3">
            {getStatusIcon()}
            <div>
              <h3 className="text-lg font-semibold text-white">{rollout?.name || 'Unknown'}</h3>
              <p className="text-sm text-slate-400">{rollout?.namespace || 'default'}</p>
            </div>
          </div>
        </div>

        <div className="space-y-3 mb-4">
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Status:</span>
            <StatusBadge status={rollout?.status || 'Unknown'} type="status" />
          </div>
          
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Strategy:</span>
            <StatusBadge status={rollout?.strategy || 'Unknown'} type="strategy" />
          </div>

          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Replicas:</span>
            <span className="text-sm text-white">{rollout?.replicas || '0/0'}</span>
          </div>

          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Age:</span>
            <span className="text-sm text-white">{rollout?.age || 'Unknown'}</span>
          </div>

          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Revision:</span>
            <span className="text-sm text-white font-mono">{rollout?.revision || '0'}</span>
          </div>
        </div>

        {/* Images Section */}
        {rollout?.images && Object.keys(rollout.images).length > 0 && (
          <div className="mb-4 p-3 bg-slate-700 rounded-lg">
            <div className="flex items-center space-x-2 mb-2">
              <PhotoIcon className="w-4 h-4 text-purple-400" />
              <span className="text-sm font-medium text-slate-300">Container Images</span>
            </div>
            <div className="space-y-3">
              {Object.entries(rollout.images).map(([container, image]) => {
                // Split by registry/path and image:tag
                const imageParts = image.split('/');
                const imageNameWithTag = imageParts[imageParts.length - 1];
                const registry = imageParts.length > 1 ? imageParts.slice(0, -1).join('/') : '';
                
                // Handle both tag and digest formats
                let imageName, version;
                if (imageNameWithTag.includes('@sha256:')) {
                  // Handle digest format: image@sha256:abc123...
                  [imageName, version] = imageNameWithTag.split('@');
                } else if (imageNameWithTag.includes(':')) {
                  // Handle tag format: image:tag
                  [imageName, version] = imageNameWithTag.split(':');
                } else {
                  imageName = imageNameWithTag;
                  version = 'latest';
                }
                
                // Truncate long SHA256 digests for display
                const displayVersion = version && version.startsWith('sha256:') 
                  ? `sha256:${version.slice(7, 19)}...` 
                  : version;
                
                return (
                  <div key={container} className="flex flex-col space-y-2">
                    <span className="text-xs text-slate-400 font-medium">{container}:</span>
                    <div className="flex flex-col space-y-1">
                      {registry && (
                        <span className="text-xs text-slate-500 font-mono break-all">
                          {registry}/
                        </span>
                      )}
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="text-xs text-white font-mono bg-slate-600 px-2 py-1 rounded break-all">
                          {imageName}
                        </span>
                        {version && (
                          <span 
                            className={`text-xs px-2 py-1 rounded font-mono break-all ${
                              version.startsWith('sha256:') 
                                ? 'text-orange-300 bg-orange-900' 
                                : 'text-purple-300 bg-purple-900'
                            }`}
                            title={version} // Show full version on hover
                          >
                            {displayVersion}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        <div className="grid grid-cols-2 gap-2 mb-3">
          <button
            onClick={() => handlePromote(false)}
            disabled={actionLoading !== null}
            className="flex items-center justify-center space-x-1 bg-green-600 hover:bg-green-700 disabled:bg-green-800 disabled:opacity-50 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
          >
            {actionLoading === 'promote' ? (
              <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
            ) : (
              <ArrowUpIcon className="w-4 h-4" />
            )}
            <span>Promote</span>
          </button>

          <button
            onClick={() => handlePromote(true)}
            disabled={actionLoading !== null}
            className="flex items-center justify-center space-x-1 bg-green-700 hover:bg-green-800 disabled:bg-green-900 disabled:opacity-50 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
          >
            {actionLoading === 'promote-full' ? (
              <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
            ) : (
              <ChevronDoubleUpIcon className="w-4 h-4" />
            )}
            <span>Full</span>
          </button>

          <button
            onClick={handlePause}
            disabled={actionLoading !== null}
            className="flex items-center justify-center space-x-1 bg-yellow-600 hover:bg-yellow-700 disabled:bg-yellow-800 disabled:opacity-50 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
          >
            {actionLoading === 'pause' ? (
              <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
            ) : (
              <PauseIcon className="w-4 h-4" />
            )}
            <span>Pause</span>
          </button>

          <button
            onClick={handleAbort}
            disabled={actionLoading !== null}
            className="flex items-center justify-center space-x-1 bg-red-600 hover:bg-red-700 disabled:bg-red-800 disabled:opacity-50 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
          >
            {actionLoading === 'abort' ? (
              <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
            ) : (
              <StopIcon className="w-4 h-4" />
            )}
            <span>Abort</span>
          </button>
        </div>

        <div className="grid grid-cols-3 gap-2">
          <button
            onClick={handleRestart}
            disabled={actionLoading !== null}
            className="flex items-center justify-center space-x-1 bg-slate-600 hover:bg-slate-700 disabled:bg-slate-800 disabled:opacity-50 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
          >
            {actionLoading === 'restart' ? (
              <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
            ) : (
              <ArrowPathIcon className="w-4 h-4" />
            )}
            <span>Restart</span>
          </button>

          <button
            onClick={() => setShowImageDialog(true)}
            disabled={actionLoading !== null}
            className="flex items-center justify-center space-x-1 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 disabled:opacity-50 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
          >
            <PhotoIcon className="w-4 h-4" />
            <span>Image</span>
          </button>

          <button
            onClick={() => setShowDetailModal(true)}
            disabled={actionLoading !== null}
            className="flex items-center justify-center space-x-1 bg-purple-600 hover:bg-purple-700 disabled:bg-purple-800 disabled:opacity-50 text-white px-3 py-2 rounded-md text-sm font-medium transition-colors"
          >
            <EyeIcon className="w-4 h-4" />
            <span>Details</span>
          </button>
        </div>
      </div>

      {/* Set Image Dialog */}
      {showImageDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-slate-800 rounded-lg p-6 w-96 border border-slate-700">
            <h3 className="text-lg font-semibold text-white mb-4">Set Image for {rollout?.name}</h3>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-2">New Image</label>
                <input
                  type="text"
                  value={newImage}
                  onChange={(e) => setNewImage(e.target.value)}
                  placeholder="nginx:1.20, myregistry/myapp:v1.2.3"
                  className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-2">Container (optional)</label>
                <input
                  type="text"
                  value={containerName}
                  onChange={(e) => setContainerName(e.target.value)}
                  placeholder="web, app, nginx"
                  className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>

            <div className="flex space-x-3 mt-6">
              <button
                onClick={handleSetImage}
                disabled={!newImage || actionLoading === 'set-image'}
                className="flex-1 flex items-center justify-center space-x-2 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 disabled:opacity-50 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors"
              >
                {actionLoading === 'set-image' ? (
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                ) : null}
                <span>Set Image</span>
              </button>
              
              <button
                onClick={() => {
                  setShowImageDialog(false);
                  setNewImage('');
                  setContainerName('');
                }}
                className="px-4 py-2 bg-slate-600 hover:bg-slate-700 text-white rounded-md text-sm font-medium transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Rollout Detail Modal */}
      <RolloutDetailModal
        rollout={rollout}
        config={config}
        isOpen={showDetailModal}
        onClose={() => setShowDetailModal(false)}
      />
    </>
  );
};

const Rollouts: React.FC = () => {
  const [rollouts, setRollouts] = useState<RolloutListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [config, setConfig] = useState<KubernetesConfig>({
    server: '',
    namespace: ''
  });
  const [error, setError] = useState<string | null>(null);
  const [showConfig, setShowConfig] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(false);

  const loadRollouts = async () => {
    // Check if Wails bindings are available
    if (!window.go || !window.go.main || !window.go.main.App) {
      setError('Wails bindings not available. Please wait for the app to fully load.');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const rolloutsData = await window.go.main.App.GetRollouts(config);
      
      if (Array.isArray(rolloutsData)) {
        setRollouts(rolloutsData);
      } else {
        setError(`Expected array, got ${typeof rolloutsData}`);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      console.error('Failed to load rollouts:', error);
      setError(`Failed to load rollouts: ${errorMessage}`);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    loadRollouts();
  };

  // Auto-refresh every 30 seconds when enabled
  useEffect(() => {
    if (autoRefresh) {
      const interval = setInterval(loadRollouts, 30000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, config]);

  // Load rollouts when config changes
  useEffect(() => {
    const checkWailsAndLoad = () => {
      if (window.go && window.go.main && window.go.main.App) {
        loadRollouts();
      } else {
        // If Wails isn't ready, wait a bit and try again
        setTimeout(checkWailsAndLoad, 100);
      }
    };
    
    checkWailsAndLoad();
  }, [config]);

  const filteredRollouts = rollouts;
  
  const statusStats = rollouts.reduce((acc, rollout) => {
    if (rollout && rollout.status) {
      const key = rollout.status.toLowerCase();
      acc[key] = (acc[key] || 0) + 1;
    }
    return acc;
  }, {} as Record<string, number>);

  return (
    <div className="min-h-screen bg-slate-900 text-white">
      <header className="bg-slate-800 border-b border-slate-700 px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <ServerIcon className="w-8 h-8 text-purple-500" />
            <div>
              <h1 className="text-xl font-bold">Yak Rollouts GUI</h1>
              <p className="text-sm text-slate-400">
                {config.namespace ? `Namespace: ${config.namespace}` : 'All namespaces'}
                {config.server && <span className="ml-2 px-2 py-1 bg-purple-600 rounded text-xs">K8s: {config.server}</span>}
              </p>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <label className="flex items-center space-x-2 text-sm">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="rounded border-slate-600 bg-slate-700 text-purple-600 focus:ring-purple-500"
              />
              <span>Auto-refresh</span>
            </label>

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
              className="flex items-center space-x-2 bg-purple-600 hover:bg-purple-700 disabled:bg-purple-800 disabled:opacity-50 px-4 py-2 rounded-md text-sm font-medium transition-colors"
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
            <h3 className="text-lg font-medium mb-4">Kubernetes Configuration</h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-2">Kubernetes Server (optional)</label>
                <input
                  type="text"
                  value={config.server}
                  onChange={(e) => setConfig({ ...config, server: e.target.value })}
                  placeholder="https://k8s.example.com"
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2">Namespace (empty = all)</label>
                <input
                  type="text"
                  value={config.namespace}
                  onChange={(e) => setConfig({ ...config, namespace: e.target.value })}
                  placeholder="default, kube-system, etc."
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-600 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                />
              </div>
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

      <main className="p-6">
        {rollouts.length > 0 && (
          <div className="grid grid-cols-4 gap-4 mb-6">
            <div className="bg-slate-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-slate-400 mb-2">Total Rollouts</h3>
              <p className="text-2xl font-bold">{rollouts.length}</p>
            </div>
            <div className="bg-slate-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-slate-400 mb-2">Healthy</h3>
              <p className="text-2xl font-bold text-green-500">{statusStats.healthy || 0}</p>
            </div>
            <div className="bg-slate-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-slate-400 mb-2">Progressing</h3>
              <p className="text-2xl font-bold text-blue-500">{statusStats.progressing || 0}</p>
            </div>
            <div className="bg-slate-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-slate-400 mb-2">Degraded</h3>
              <p className="text-2xl font-bold text-red-500">{statusStats.degraded || 0}</p>
            </div>
          </div>
        )}

        {loading && rollouts.length === 0 ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="w-8 h-8 border-2 border-purple-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
              <p className="text-slate-400">Loading rollouts...</p>
            </div>
          </div>
        ) : filteredRollouts.length === 0 ? (
          <div className="text-center py-12">
            <ServerIcon className="w-12 h-12 text-slate-500 mx-auto mb-4" />
            <p className="text-slate-400">No rollouts found</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredRollouts.map((rollout, index) => (
              <RolloutCard
                key={rollout?.name || `rollout-${index}`}
                rollout={rollout}
                config={config}
                onAction={handleRefresh}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  );
};

export default Rollouts;