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
  `status` enum('pending','active','disabled','completed') NOT NULL DEFAULT 'pending',
  `showStatus` enum('site','direct','no') NOT NULL DEFAULT 'site',
  `signupStatus` enum('open','with_code','closed') NOT NULL DEFAULT 'open',
  `maxParticipants` int(6) NOT NULL DEFAULT 0,
  `participantVisibility` enum('code','email','full') NOT NULL DEFAULT 'code',
  `participantMinimumAge` int(3) NOT NULL DEFAULT 0,
  `connectParticipantToConsentForm` enum('yes','no') NOT NULL DEFAULT 'yes',
  `completeMessage` varchar(5096) NOT NULL DEFAULT '',
  `flowRule` enum('free','in_order_in_module','in_order_for_project') NOT NULL DEFAULT 'free',
  `completeRule` enum('continued_access','blocked') NOT NULL DEFAULT 'continued_access',
  `startRule` enum('any','date','threshold') NOT NULL DEFAULT 'any',
  `startDate` datetime NOT NULL,
  `endDate` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `siteId` (`siteId`),
  KEY `status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4;


DROP TABLE IF EXISTS `ProjectUserLinks`;
CREATE TABLE `ProjectUserLinks` (
  `projectId` int(11) NOT NULL,
  `userId` int(11) NOT NULL,
  `status` enum('not_started', 'started', 'completed') NOT NULL DEFAULT 'not_started',
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
  `summary` varchar(2048) NOT NULL,
  `blockType` enum('other','form','embed', 'text', 'external', 'file') NOT NULL DEFAULT 'other',
  `allowReset` enum('yes', 'no') NOT NULL DEFAULT 'yes',
  PRIMARY KEY (`id`),
  KEY `blockType` (`blockType`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `BlockModuleFlows`;
CREATE TABLE `BlockModuleFlows` (
  `blockId` int(11) NOT NULL,
  `moduleId` int(11) NOT NULL,
  `flowOrder` int(11) NOT NULL,
  PRIMARY KEY (`blockId`, `moduleId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


DROP TABLE IF EXISTS `BlockForm`;
CREATE TABLE `BlockForm` (
  `blockId` int(11) NOT NULL,
  `formType` ENUM('survey', 'quiz') NOT NULL DEFAULT 'survey',
  PRIMARY KEY (`blockId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `BlockFormQuestions`;
CREATE TABLE `BlockFormQuestions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `blockId` int(11) NOT NULL,
  `questionType` enum('explanation', 'multiple','single','short','long', 'likert5', 'likert7') DEFAULT 'explanation',
  `question` varchar(2048) NOT NULL DEFAULT '',
  `formOrder` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY (`blockId`),
  KEY (`formOrder`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `BlockFormQuestionOptions`;
CREATE TABLE `BlockFormQuestionOptions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `questionId` int(11) NOT NULL,
  `optionText` varchar(1024) NOT NULL DEFAULT '',
  `optionOrder` int(11) DEFAULT NULL,
  `optionIsCorrect` ENUM('na', 'yes', 'no') NOT NULL DEFAULT 'na',
  PRIMARY KEY (`id`),
  KEY (`questionId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `BlockFormSubmissions`;
CREATE TABLE `BlockFormSubmissions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `blockId` int(11) NOT NULL,
  `userId` int(11) NOT NULL,
  `submittedOn` datetime NOT NULL,
  `results` enum('na', 'needs_input', 'passed', 'failed'),
  PRIMARY KEY (`id`),
  KEY (`blockId`),
  KEY (`userId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `BlockFormSubmissionResponses`;
CREATE TABLE `BlockFormSubmissionResponses` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `submissionId` int(11) NOT NULL,
  `questionId` int(11) NOT NULL,
  `optionId` int(11) NOT NULL,
  `textResponse` varchar(2048) NOT NULL default '',
  `isCorrect` enum('na', 'pending', 'yes', 'no'),
  PRIMARY KEY (`id`),
  KEY (`submissionId`),
  KEY (`optionId`),
  KEY (`questionId`)
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
  `projectId` int(11) NOT NULL,
  `contentInMarkdown` text NOT NULL,
  `contactInformationDisplay` varchar(2048) NOT NULL DEFAULT '',
  `institutionInformationDisplay` varchar(2048) NOT NULL DEFAULT '',
  PRIMARY KEY (`projectId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `ConsentResponses`;
CREATE TABLE `ConsentResponses` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `projectId` int(11) NOT NULL,
  `submittedOn` datetime NOT NULL,
  `consentStatus` enum('accepted','accepted_for_other','declined') DEFAULT NULL,
  `participantComments` text NOT NULL,
  `researcherComments` text NOT NULL,
  `participantProvidedFirstName` varchar(64) NOT NULL,
  `participantProvidedLastName` varchar(64) NOT NULL,
  `participantProvidedContactInformation` varchar(64) NOT NULL,
  `participantId` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `projectId` (`projectId`),
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

DROP TABLE IF EXISTS `BlockText`;
CREATE TABLE `BlockText` (
  `blockId` int(11) NOT NULL,
  `text` text NOT NULL,
  PRIMARY KEY (`blockId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `BlockExternal`;
CREATE TABLE `BlockExternal` (
  `blockId` int(11) NOT NULL,
  `externalLink` varchar(2048) NOT NULL,
  PRIMARY KEY (`blockId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `BlockEmbed`;
CREATE TABLE `BlockEmbed` (
  `blockId` int(11) NOT NULL,
  `embedType` enum('youtube','external_pdf', 'internal_pdf') NOT NULL DEFAULT 'external_pdf',
  `embedLink` varchar(2048) NOT NULL,
  `fileId` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`blockId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `Files`;
CREATE TABLE `Files` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `remoteKey` varchar(512) NOT NULL DEFAULT '',
  `display` varchar(512) NOT NULL DEFAULT '',
  `description` varchar(2048) NOT NULL DEFAULT '',
  `fileType` varchar(32) NOT NULL DEFAULT '',
  `uploadedOn` datetime NOT NULL DEFAULT current_timestamp(),
  `uploadedBy` int(11) NOT NULL DEFAULT 0,
  `visibility` enum('admin', 'users', 'project','public') NOT NULL DEFAULT 'admin',
  `fileSize` int(11) NOT NULL DEFAULT 0,
  `locationSource` enum('other', 'aws') NOT NULL DEFAULT 'other',
  PRIMARY KEY (`id`),
  KEY `visibility` (`visibility`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `BlockFile`;
CREATE TABLE `BlockFile` (
  `blockId` int(11) NOT NULL,
  `fileId` int(11) NOT NULL,
  PRIMARY KEY (`blockId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `Notes` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `createdOn` datetime NOT NULL,
  `noteType` enum('journal','project') NOT NULL DEFAULT 'journal',
  `projectId` int(11) NOT NULL DEFAULT 0,
  `moduleId` int(11) NOT NULL DEFAULT 0,
  `blockId` int(11) NOT NULL DEFAULT 0,
  `visibility` enum('private','admins') NOT NULL DEFAULT 'private',
  `title` varchar(512) NOT NULL DEFAULT '',
  `body` varchar(5096) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `userId` (`userId`),
  KEY `projectId` (`projectId`),
  KEY `ids` (`userId`,`projectId`,`moduleId`,`blockId`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4;