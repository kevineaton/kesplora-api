DROP TABLE IF EXISTS `Site`;
CREATE TABLE `Site` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `createdOn` datetime not null,
  `shortName` varchar(64) NOT NULL,
  `name` varchar(256) NOT NULL,
  `description` text NOT NULL,
  `domain` varchar(256) NOT NULL,
  `status` enum('pending','active','disabled') NOT NULL DEFAULT 'pending',
  `projectListOptions` enum('show_all', 'show_active', 'show_none') NOT NULL DEFAULT 'show_active',
  `siteTechnicalContact` varchar(2048) not null default '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `Projects`;
CREATE TABLE `Projects` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `siteId` int(11) NOT NULL,
  `name` varchar(128) NOT NULL,
  `shortCode` varchar(16) NOT NULL DEFAULT '',
  `shortDescription` varchar(1024) NOT NULL DEFAULT '',
  `description` text NOT NULL,
  `status` enum('pending','active','disabled', 'completed') NOT NULL DEFAULT 'pending',
  `showStatus` enum('site','direct','no') NOT NULL DEFAULT 'site',
  `signupStatus` enum('open','with_code','closed') NOT NULL DEFAULT 'open',
  `maxParticipants` int(6) NOT NULL DEFAULT 0,
  `participantVisibility` enum('code','email','full') NOT NULL DEFAULT 'code',
  `participantMinimumAge` int(3) NOT NULL DEFAULT 0,
  `connectParticipantToConsentForm` enum('yes','no') NOT NULL DEFAULT 'yes',
  PRIMARY KEY (`id`),
  KEY `siteId` (`siteId`),
  KEY `status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


DROP TABLE IF EXISTS `ProjectUserLinks`;
CREATE TABLE `ProjectUserLinks` (
  `projectId` int(11) NOT NULL,
  `userId` int(11) NOT NULL,
  PRIMARY KEY (`projectId`, `userId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `Flows`;
CREATE TABLE `Flows` (
  `projectId` int(11) NOT NULL,
  `moduleId` int(11) NOT NULL,
  `flowOrder` int(11) NOT NULL,
  PRIMARY KEY (`projectId`, `moduleId`),
  KEY `projectId` (`projectId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `Modules`;
CREATE TABLE `Modules` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `status` enum('pending','active','disabled') DEFAULT 'pending',
  `description` varchar(5096) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `Blocks`;
CREATE TABLE `Blocks` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `blockType` enum('other','sign_up','survey','presentation') NOT NULL DEFAULT 'other',
  `blockTypeId` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `moduleId` (`moduleId`),
  KEY `blockType` (`blockType`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `BlockModuleFlows`;
CREATE TABLE `BlockModuleFlows` (
  `blockId` int(11) NOT NULL,
  `moduleId` int(11) NOT NULL,
  `flowOrder` int(11) NOT NULL,
  PRIMARY KEY (`blockId`, `moduleId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


DROP TABLE IF EXISTS `BlockUserStatus`;
CREATE TABLE `BlockUserStatus` (
  `userId` int(11) NOT NULL,
  `blockId` int(11) NOT NULL,
  `moduleId` int(11) NOT NULL,
  `projectId` int(11) NOT NULL,
  `lastUpdatedOn` datetime NOT NULL,
  `status` ENUM('not_started', 'started', 'completed') NOT NULL DEFAULT 'not_started',
  PRIMARY KEY (`userId`, `blockId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `Users`;
CREATE TABLE `Users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(32) NOT NULL DEFAULT '',
  `firstName` varchar(64) NOT NULL DEFAULT '',
  `lastName` varchar(64) NOT NULL DEFAULT '',
  `pronouns` varchar(32) NOT NULL DEFAULT '',
  `email` varchar(256) NOT NULL DEFAULT '',
  `password` varchar(64) NOT NULL,
  `dateOfBirth` date NOT NULL DEFAULT '1970-01-01',
  `participantCode` varchar(32) NOT NULL DEFAULT '',
  `status` enum('active','pending','locked','disabled') NOT NULL DEFAULT 'pending',
  `systemRole` enum('user','admin', 'participant') NOT NULL DEFAULT 'user',
  `createdOn` datetime NOT NULL,
  `lastLoginOn` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `email` (`email`),
  KEY `participantCode` (`participantCode`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ;

DROP TABLE IF EXISTS `ConsentForms`;
CREATE TABLE `ConsentForms` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `projectId` int(11) NOT NULL,
  `contentInMarkdown` text NOT NULL,
  `contactInformationDisplay` varchar(512) NOT NULL DEFAULT '',
  `institutionInformationDisplay` varchar(512) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `projectId` (`projectId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `ConsentResponses`;
CREATE TABLE `ConsentResponses` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `consentFormId` int(11) NOT NULL,
  `dateConsented` datetime NOT NULL,
  `consentStatus` enum('accepted','accepted_for_other','declined') DEFAULT NULL,
  `participantComments` text NOT NULL,
  `researcherComments` text NOT NULL,
  `participantProvidedFirstName` varchar(64) NOT NULL,
  `participantProvidedLastName` varchar(64) NOT NULL,
  `participantProvidedContactInformation` varchar(64) NOT NULL,
  `participantId` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `consentFormId` (`consentFormId`),
  KEY `participantId` (`participantId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `Tokens`;
CREATE TABLE `Tokens` (
  `userId` int(11) NOT NULL,
  `tokenType` enum('email','password_reset','refresh') NOT NULL DEFAULT 'email',
  `createdOn` datetime NOT NULL,
  `expiresOn` datetime NOT NULL,
  `token` varchar(128) NOT NULL,
  UNIQUE KEY `userId` (`userId`,`tokenType`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;