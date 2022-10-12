set names utf8mb4;

use n9e_v5;

CREATE TABLE `ltw_host_ctf` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `ip` varchar(64) NOT NULL COMMENT 'ip',
    `hostname` varchar(64) NOT NULL COMMENT '主机名',
    `status` varchar(64) NOT NULL COMMENT '状态',
    `create_at` bigint(20) NOT NULL DEFAULT '0',
    `create_by` varchar(64) NOT NULL DEFAULT '',
    `update_at` bigint(20) NOT NULL DEFAULT '0',
    `update_by` varchar(64) NOT NULL DEFAULT '',
    `version` varchar(20) DEFAULT NULL,
    `actions` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `ip` (`ip`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4;

CREATE TABLE `ltw_host_ctf_conf` (
    `id` bigint unsigned not null auto_increment,
    `host_ctf_id` bigint unsigned,
    `ip` varchar(64) not null comment 'ip',
    `name` varchar(64) not null comment 'ctf config name',
    `status` varchar(64) not null comment '状态',
    `local_toml` varchar(5000) comment '本地配置文件',
    `remote_toml` varchar(5000) comment '服务器配置文件',
    `create_at` bigint not null default 0,
    `create_by` varchar(64) not null default '',
    `update_at` bigint not null default 0,
    `update_by` varchar(64) not null default '',
    PRIMARY KEY (`id`),
    KEY (`ip`),
    KEY (`name`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE `ltw_host_ctf_conf_index` (
    `id` bigint unsigned not null auto_increment,
    `host_ctf_conf_id` bigint unsigned,
    `name` varchar(64) not null,
    `index` varchar(64) not null,
    `value` varchar(255) default '',
    `create_at` bigint not null default 0,
    `create_by` varchar(64) not null default '',
    `update_at` bigint not null default 0,
    `update_by` varchar(64) not null default '',
    PRIMARY KEY (`id`),
    KEY (`host_ctf_conf_id`),
    KEY (`index`, `value`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE `ltw_host_ctf_conf_log` (
    `id` bigint unsigned not null auto_increment,
    `ip` varchar(64) default 'ip地址',
    `hostname` varchar(64) default '主机名',
    `host_ctf_conf_id` bigint unsigned,
    `name` varchar(64) not null,
    `stand_out` text,
    `status` varchar(255) default '任务执行状态',
    `message` varchar(255) default '日志信息',
    `last_toml` varchar(5000) default '上一次配置',
    `current_toml` varchar(5000) default '当前配置',
    `create_at` bigint not null default 0,
    `create_by` varchar(64) not null default '',
    `update_at` bigint not null default 0,
    `update_by` varchar(64) not null default '',
    PRIMARY KEY (`id`),
    KEY (`ip`),
    KEY (`name`),
    KEY (`name`, `ip`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE `ltw_duty_conf` (
    `id` bigint unsigned not null auto_increment,
    `busy_group_id` bigint unsigned,
    `name` varchar(64),
    `start_time` varchar(64) default '00:00',
    `end_time` varchar(64) default '24:00',
    `priority` tinyint unsigned,
    `create_at` bigint not null default 0,
    `create_by` varchar(64) not null default '',
    `update_at` bigint not null default 0,
    `update_by` varchar(64) not null default '',
    PRIMARY KEY (`id`),
    KEY (`name`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
INSERT INTO `ltw_duty_conf` (`id`, `busy_group_id`, `name`, `start_time`, `end_time`, `priority`, `create_at`, `create_by`, `update_at`, `update_by`) VALUES (1, 1, '整天', '08:00', '08:00', 1, 0, '', 0, '');
INSERT INTO `ltw_duty_conf` (`id`, `busy_group_id`, `name`, `start_time`, `end_time`, `priority`, `create_at`, `create_by`, `update_at`, `update_by`) VALUES (2, 1, '夜班', '18:00', '8:00', 2, 0, '', 0, '');
INSERT INTO `ltw_duty_conf` (`id`, `busy_group_id`, `name`, `start_time`, `end_time`, `priority`, `create_at`, `create_by`, `update_at`, `update_by`) VALUES (3, 1, '早班', '08:00', '10:00', 3, 0, '', 0, '');
INSERT INTO `ltw_duty_conf` (`id`, `busy_group_id`, `name`, `start_time`, `end_time`, `priority`, `create_at`, `create_by`, `update_at`, `update_by`) VALUES (4, 1, '午班', '10:00', '18:00', 3, 0, '', 0, '');
INSERT INTO `ltw_duty_conf` (`id`, `busy_group_id`, `name`, `start_time`, `end_time`, `priority`, `create_at`, `create_by`, `update_at`, `update_by`) VALUES (5, 2, '整天', '08:00', '08:00', 1, 0, '', 0, '');
INSERT INTO `ltw_duty_conf` (`id`, `busy_group_id`, `name`, `start_time`, `end_time`, `priority`, `create_at`, `create_by`, `update_at`, `update_by`) VALUES (6, 2, '白班', '08:00', '18:00', 2, 0, '', 0, '');

CREATE TABLE `ltw_duty_roster` (
    `id` bigint unsigned not null auto_increment,
    `duty_conf_id` bigint unsigned,
    `user_id` bigint unsigned,
    `start_at` bigint not null default 0,
    `end_at` bigint not null default 0,
    `create_at` bigint not null default 0,
    `create_by` varchar(64) not null default '',
    `update_at` bigint not null default 0,
    `update_by` varchar(64) not null default '',
    PRIMARY KEY (`id`),
    KEY (`user_id`, `start_at`, `end_at`),
    KEY (`duty_conf_id`),
    KEY (`user_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1001, 'orch_国内_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1002, 'orch_微软_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1003, 'orch_平安办_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1004, 'orch_香港_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1005, 'orch_国电投_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1006, 'orch_中交_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1007, 'orch_中国电子_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1008, 'orch_中能建_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1009, 'orch_中电科_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1010, 'orch_万达_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1011, 'orch_中烟_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1012, 'orch_有研_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1013, 'orch_有矿_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1014, 'orch_中交海外组网_告警', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');
