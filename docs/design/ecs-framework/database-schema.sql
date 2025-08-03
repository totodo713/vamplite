-- =============================================
-- ECS Framework Database Schema
-- Muscle Dreamer Game Engine
-- Purpose: Save/Load Game State, MOD Data, Performance Metrics
-- Generated: 2025-08-03
-- =============================================

-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- =============================================
-- 1. Game Save System Tables
-- =============================================

-- Save slots management
CREATE TABLE save_slots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slot_number INTEGER NOT NULL UNIQUE CHECK (slot_number >= 1 AND slot_number <= 10),
    save_name TEXT NOT NULL,
    description TEXT,
    thumbnail_path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    game_version TEXT NOT NULL,
    world_seed INTEGER,
    play_time_seconds INTEGER DEFAULT 0,
    is_auto_save BOOLEAN DEFAULT FALSE,
    metadata TEXT -- JSON metadata
);

CREATE INDEX idx_save_slots_slot_number ON save_slots(slot_number);
CREATE INDEX idx_save_slots_created_at ON save_slots(created_at);
CREATE INDEX idx_save_slots_auto_save ON save_slots(is_auto_save);

-- World state snapshots
CREATE TABLE world_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    save_slot_id INTEGER NOT NULL,
    snapshot_type TEXT NOT NULL CHECK (snapshot_type IN ('full', 'incremental', 'checkpoint')),
    world_state TEXT NOT NULL, -- JSON serialized world state
    entity_count INTEGER NOT NULL DEFAULT 0,
    component_count INTEGER NOT NULL DEFAULT 0,
    compressed_size INTEGER NOT NULL DEFAULT 0,
    checksum TEXT, -- SHA256 hash for integrity
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (save_slot_id) REFERENCES save_slots(id) ON DELETE CASCADE
);

CREATE INDEX idx_world_snapshots_save_slot ON world_snapshots(save_slot_id);
CREATE INDEX idx_world_snapshots_type ON world_snapshots(snapshot_type);
CREATE INDEX idx_world_snapshots_created_at ON world_snapshots(created_at);

-- Entity persistence
CREATE TABLE saved_entities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    world_snapshot_id INTEGER NOT NULL,
    entity_id INTEGER NOT NULL,
    entity_type TEXT, -- Optional entity classification
    is_persistent BOOLEAN DEFAULT TRUE,
    parent_entity_id INTEGER, -- For hierarchical entities
    entity_data TEXT NOT NULL, -- JSON serialized entity data
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (world_snapshot_id) REFERENCES world_snapshots(id) ON DELETE CASCADE
);

CREATE INDEX idx_saved_entities_snapshot ON saved_entities(world_snapshot_id);
CREATE INDEX idx_saved_entities_entity_id ON saved_entities(entity_id);
CREATE INDEX idx_saved_entities_parent ON saved_entities(parent_entity_id);
CREATE INDEX idx_saved_entities_type ON saved_entities(entity_type);

-- Component persistence
CREATE TABLE saved_components (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    saved_entity_id INTEGER NOT NULL,
    component_type TEXT NOT NULL,
    component_data TEXT NOT NULL, -- JSON serialized component data
    component_version INTEGER DEFAULT 1,
    is_dirty BOOLEAN DEFAULT FALSE,
    
    FOREIGN KEY (saved_entity_id) REFERENCES saved_entities(id) ON DELETE CASCADE
);

CREATE INDEX idx_saved_components_entity ON saved_components(saved_entity_id);
CREATE INDEX idx_saved_components_type ON saved_components(component_type);
CREATE INDEX idx_saved_components_dirty ON saved_components(is_dirty);

-- =============================================
-- 2. MOD System Tables
-- =============================================

-- MOD registry
CREATE TABLE mods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    mod_id TEXT NOT NULL UNIQUE, -- Unique MOD identifier
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    author TEXT,
    description TEXT,
    homepage_url TEXT,
    repository_url TEXT,
    license TEXT,
    
    -- Installation info
    install_path TEXT NOT NULL,
    installed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- State
    is_enabled BOOLEAN DEFAULT FALSE,
    is_validated BOOLEAN DEFAULT FALSE,
    
    -- Compatibility
    game_version_min TEXT,
    game_version_max TEXT,
    api_version TEXT,
    
    -- Metadata
    metadata TEXT, -- JSON additional metadata
    checksum TEXT  -- MOD file integrity hash
);

CREATE INDEX idx_mods_mod_id ON mods(mod_id);
CREATE INDEX idx_mods_enabled ON mods(is_enabled);
CREATE INDEX idx_mods_validated ON mods(is_validated);

-- MOD dependencies
CREATE TABLE mod_dependencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    mod_id INTEGER NOT NULL,
    dependency_mod_id TEXT NOT NULL,
    dependency_version_min TEXT,
    dependency_version_max TEXT,
    is_optional BOOLEAN DEFAULT FALSE,
    
    FOREIGN KEY (mod_id) REFERENCES mods(id) ON DELETE CASCADE
);

CREATE INDEX idx_mod_dependencies_mod ON mod_dependencies(mod_id);
CREATE INDEX idx_mod_dependencies_dependency ON mod_dependencies(dependency_mod_id);

-- MOD permissions
CREATE TABLE mod_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    mod_id INTEGER NOT NULL,
    permission_type TEXT NOT NULL,
    permission_value TEXT,
    granted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (mod_id) REFERENCES mods(id) ON DELETE CASCADE
);

CREATE INDEX idx_mod_permissions_mod ON mod_permissions(mod_id);
CREATE INDEX idx_mod_permissions_type ON mod_permissions(permission_type);

-- MOD resource usage tracking
CREATE TABLE mod_resource_usage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    mod_id INTEGER NOT NULL,
    resource_type TEXT NOT NULL, -- 'memory', 'cpu_time', 'entity_count', etc.
    current_usage REAL NOT NULL DEFAULT 0.0,
    peak_usage REAL NOT NULL DEFAULT 0.0,
    limit_value REAL,
    last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (mod_id) REFERENCES mods(id) ON DELETE CASCADE
);

CREATE INDEX idx_mod_resource_usage_mod ON mod_resource_usage(mod_id);
CREATE INDEX idx_mod_resource_usage_type ON mod_resource_usage(resource_type);

-- =============================================
-- 3. Performance Metrics Tables
-- =============================================

-- System performance history
CREATE TABLE performance_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL, -- Game session identifier
    
    -- Timing metrics (microseconds)
    frame_time_us INTEGER NOT NULL,
    update_time_us INTEGER NOT NULL,
    render_time_us INTEGER NOT NULL,
    gc_time_us INTEGER DEFAULT 0,
    
    -- Entity/Component metrics
    entity_count INTEGER NOT NULL DEFAULT 0,
    component_count INTEGER NOT NULL DEFAULT 0,
    system_count INTEGER NOT NULL DEFAULT 0,
    active_queries INTEGER DEFAULT 0,
    
    -- Memory metrics (bytes)
    memory_used INTEGER NOT NULL DEFAULT 0,
    memory_allocated INTEGER NOT NULL DEFAULT 0,
    memory_peak INTEGER DEFAULT 0,
    
    -- Game state
    current_level TEXT,
    player_count INTEGER DEFAULT 1,
    
    -- Timestamp
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_performance_metrics_session ON performance_metrics(session_id);
CREATE INDEX idx_performance_metrics_recorded_at ON performance_metrics(recorded_at);
CREATE INDEX idx_performance_metrics_frame_time ON performance_metrics(frame_time_us);

-- System-specific performance metrics
CREATE TABLE system_performance (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    performance_metric_id INTEGER NOT NULL,
    system_type TEXT NOT NULL,
    system_name TEXT,
    
    -- System timing (microseconds)
    execution_time_us INTEGER NOT NULL,
    entities_processed INTEGER DEFAULT 0,
    
    -- System state
    is_enabled BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    
    FOREIGN KEY (performance_metric_id) REFERENCES performance_metrics(id) ON DELETE CASCADE
);

CREATE INDEX idx_system_performance_metric ON system_performance(performance_metric_id);
CREATE INDEX idx_system_performance_type ON system_performance(system_type);
CREATE INDEX idx_system_performance_execution_time ON system_performance(execution_time_us);

-- Query performance tracking
CREATE TABLE query_performance (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    query_signature TEXT NOT NULL,
    
    -- Query metrics
    execution_count INTEGER NOT NULL DEFAULT 1,
    total_execution_time_us INTEGER NOT NULL,
    average_execution_time_us REAL NOT NULL,
    min_execution_time_us INTEGER NOT NULL,
    max_execution_time_us INTEGER NOT NULL,
    
    -- Query characteristics
    component_types TEXT, -- JSON array of component types
    entity_count_average REAL DEFAULT 0.0,
    cache_hit_rate REAL DEFAULT 0.0,
    
    -- Tracking
    first_executed DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_executed DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_updated DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_query_performance_signature ON query_performance(query_signature);
CREATE INDEX idx_query_performance_last_executed ON query_performance(last_executed);
CREATE INDEX idx_query_performance_avg_time ON query_performance(average_execution_time_us);

-- Performance alerts/thresholds
CREATE TABLE performance_alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    alert_type TEXT NOT NULL, -- 'threshold_exceeded', 'performance_degradation', etc.
    severity TEXT NOT NULL CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    metric_type TEXT NOT NULL,
    threshold_value REAL,
    actual_value REAL NOT NULL,
    message TEXT NOT NULL,
    
    -- Context
    session_id TEXT,
    system_type TEXT,
    
    -- Status
    is_resolved BOOLEAN DEFAULT FALSE,
    resolved_at DATETIME,
    
    -- Timing
    triggered_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_performance_alerts_type ON performance_alerts(alert_type);
CREATE INDEX idx_performance_alerts_severity ON performance_alerts(severity);
CREATE INDEX idx_performance_alerts_resolved ON performance_alerts(is_resolved);
CREATE INDEX idx_performance_alerts_triggered_at ON performance_alerts(triggered_at);

-- =============================================
-- 4. Configuration Tables
-- =============================================

-- ECS configuration settings
CREATE TABLE ecs_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    config_key TEXT NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    config_type TEXT NOT NULL CHECK (config_type IN ('string', 'integer', 'float', 'boolean', 'json')),
    description TEXT,
    is_runtime_modifiable BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ecs_config_key ON ecs_config(config_key);
CREATE INDEX idx_ecs_config_runtime_modifiable ON ecs_config(is_runtime_modifiable);

-- Insert default ECS configuration
INSERT INTO ecs_config (config_key, config_value, config_type, description, is_runtime_modifiable) VALUES
('max_entities', '10000', 'integer', 'Maximum number of entities allowed', FALSE),
('memory_limit_mb', '256', 'integer', 'Memory limit in megabytes', FALSE),
('enable_metrics', 'true', 'boolean', 'Enable performance metrics collection', TRUE),
('enable_events', 'true', 'boolean', 'Enable event system', TRUE),
('thread_pool_size', '4', 'integer', 'Thread pool size for parallel systems', FALSE),
('query_cache_size', '1000', 'integer', 'Maximum number of cached queries', TRUE),
('gc_interval_seconds', '30', 'integer', 'Garbage collection interval in seconds', TRUE),
('enable_debug_logging', 'false', 'boolean', 'Enable debug-level logging', TRUE),
('auto_save_interval_minutes', '5', 'integer', 'Auto-save interval in minutes', TRUE),
('max_save_slots', '10', 'integer', 'Maximum number of save slots', FALSE);

-- Component type registry
CREATE TABLE component_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    component_type TEXT NOT NULL UNIQUE,
    component_name TEXT NOT NULL,
    description TEXT,
    schema_version INTEGER DEFAULT 1,
    is_serializable BOOLEAN DEFAULT TRUE,
    is_networked BOOLEAN DEFAULT FALSE,
    size_hint INTEGER, -- Expected size in bytes
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert core component types
INSERT INTO component_types (component_type, component_name, description, is_serializable, is_networked, size_hint) VALUES
('transform', 'Transform Component', 'Position, rotation, and scale data', TRUE, TRUE, 64),
('sprite', 'Sprite Component', 'Sprite rendering data', TRUE, FALSE, 128),
('physics', 'Physics Component', 'Physics properties and state', TRUE, TRUE, 96),
('health', 'Health Component', 'Health and damage tracking', TRUE, TRUE, 48),
('ai', 'AI Component', 'AI behavior and state', TRUE, FALSE, 256),
('inventory', 'Inventory Component', 'Item container and management', TRUE, FALSE, 512),
('audio', 'Audio Component', 'Audio source and properties', FALSE, FALSE, 32),
('input', 'Input Component', 'Input handling configuration', FALSE, FALSE, 64);

-- System type registry
CREATE TABLE system_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    system_type TEXT NOT NULL UNIQUE,
    system_name TEXT NOT NULL,
    description TEXT,
    category TEXT, -- 'update', 'render', 'input', 'physics', etc.
    priority INTEGER DEFAULT 50,
    is_thread_safe BOOLEAN DEFAULT FALSE,
    required_components TEXT, -- JSON array of required component types
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert core system types
INSERT INTO system_types (system_type, system_name, description, category, priority, is_thread_safe, required_components) VALUES
('input', 'Input System', 'Handles player input processing', 'input', 90, FALSE, '["input"]'),
('ai', 'AI System', 'Processes AI behavior', 'logic', 80, TRUE, '["ai", "transform"]'),
('physics', 'Physics System', 'Physics simulation and movement', 'physics', 70, TRUE, '["physics", "transform"]'),
('collision', 'Collision System', 'Collision detection and response', 'physics', 60, TRUE, '["transform"]'),
('animation', 'Animation System', 'Sprite animation processing', 'rendering', 50, TRUE, '["sprite"]'),
('audio', 'Audio System', 'Audio playback and management', 'audio', 40, TRUE, '["audio", "transform"]'),
('rendering', 'Rendering System', 'Sprite and UI rendering', 'rendering', 30, FALSE, '["sprite", "transform"]'),
('ui', 'UI System', 'User interface rendering', 'rendering', 20, FALSE, '[]'),
('debug', 'Debug System', 'Debug information display', 'debug', 10, FALSE, '[]');

-- =============================================
-- 5. Utility Views
-- =============================================

-- Performance summary view
CREATE VIEW performance_summary AS
SELECT 
    session_id,
    COUNT(*) as metric_count,
    AVG(frame_time_us) as avg_frame_time_us,
    MAX(frame_time_us) as max_frame_time_us,
    AVG(entity_count) as avg_entity_count,
    MAX(entity_count) as max_entity_count,
    AVG(memory_used) as avg_memory_used,
    MAX(memory_used) as max_memory_used,
    MIN(recorded_at) as session_start,
    MAX(recorded_at) as session_end
FROM performance_metrics 
GROUP BY session_id;

-- Active MODs view
CREATE VIEW active_mods AS
SELECT 
    m.mod_id,
    m.name,
    m.version,
    m.author,
    m.installed_at,
    COUNT(mp.id) as permission_count,
    AVG(mru.current_usage) as avg_resource_usage
FROM mods m
LEFT JOIN mod_permissions mp ON m.id = mp.mod_id
LEFT JOIN mod_resource_usage mru ON m.id = mru.mod_id
WHERE m.is_enabled = TRUE AND m.is_validated = TRUE
GROUP BY m.id;

-- System performance summary view
CREATE VIEW system_performance_summary AS
SELECT 
    system_type,
    system_name,
    COUNT(*) as execution_count,
    AVG(execution_time_us) as avg_execution_time_us,
    MAX(execution_time_us) as max_execution_time_us,
    AVG(entities_processed) as avg_entities_processed,
    MAX(entities_processed) as max_entities_processed
FROM system_performance sp
JOIN system_types st ON sp.system_type = st.system_type
GROUP BY sp.system_type;

-- Recent save slots view
CREATE VIEW recent_saves AS
SELECT 
    ss.slot_number,
    ss.save_name,
    ss.description,
    ss.updated_at,
    ss.play_time_seconds,
    ws.entity_count,
    ws.component_count
FROM save_slots ss
LEFT JOIN world_snapshots ws ON ss.id = ws.save_slot_id 
    AND ws.id = (
        SELECT MAX(id) FROM world_snapshots WHERE save_slot_id = ss.id
    )
ORDER BY ss.updated_at DESC;

-- =============================================
-- 6. Triggers for Data Integrity
-- =============================================

-- Update save_slots.updated_at when world_snapshots are inserted
CREATE TRIGGER update_save_slot_timestamp
    AFTER INSERT ON world_snapshots
BEGIN
    UPDATE save_slots 
    SET updated_at = CURRENT_TIMESTAMP 
    WHERE id = NEW.save_slot_id;
END;

-- Validate MOD dependencies on insert
CREATE TRIGGER validate_mod_dependency
    BEFORE INSERT ON mod_dependencies
BEGIN
    SELECT CASE
        WHEN (SELECT COUNT(*) FROM mods WHERE mod_id = NEW.dependency_mod_id) = 0
        THEN RAISE(ABORT, 'Dependency MOD does not exist')
    END;
END;

-- Clean up old performance metrics (keep only last 30 days)
CREATE TRIGGER cleanup_old_metrics
    AFTER INSERT ON performance_metrics
BEGIN
    DELETE FROM performance_metrics 
    WHERE recorded_at < datetime('now', '-30 days')
    AND id NOT IN (
        SELECT id FROM performance_metrics 
        ORDER BY recorded_at DESC 
        LIMIT 10000
    );
END;

-- =============================================
-- 7. Indexes for Performance
-- =============================================

-- Composite indexes for common queries
CREATE INDEX idx_saved_entities_snapshot_type ON saved_entities(world_snapshot_id, entity_type);
CREATE INDEX idx_saved_components_entity_type ON saved_components(saved_entity_id, component_type);
CREATE INDEX idx_performance_metrics_session_recorded ON performance_metrics(session_id, recorded_at);
CREATE INDEX idx_system_performance_type_time ON system_performance(system_type, execution_time_us);

-- =============================================
-- 8. Database Maintenance Procedures
-- =============================================

/*
-- Maintenance SQL commands (run periodically)

-- Vacuum database to reclaim space
VACUUM;

-- Analyze tables for query optimization
ANALYZE;

-- Clean up old auto-saves (keep only latest 5 per slot)
DELETE FROM world_snapshots 
WHERE save_slot_id IN (
    SELECT id FROM save_slots WHERE is_auto_save = TRUE
) 
AND id NOT IN (
    SELECT id FROM world_snapshots ws
    WHERE save_slot_id IN (SELECT id FROM save_slots WHERE is_auto_save = TRUE)
    ORDER BY created_at DESC
    LIMIT 5
);

-- Update query performance statistics
UPDATE query_performance 
SET last_updated = CURRENT_TIMESTAMP
WHERE last_executed < datetime('now', '-1 hour');

-- Archive old performance data
INSERT INTO performance_metrics_archive 
SELECT * FROM performance_metrics 
WHERE recorded_at < datetime('now', '-90 days');

DELETE FROM performance_metrics 
WHERE recorded_at < datetime('now', '-90 days');
*/