CREATE TABLE `Site` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `shortName` varchar(64) NOT NULL,
  `name` varchar(256) NOT NULL,
  `description` text NOT NULL,
  `domain` varchar(256) NOT NULL,
  `status` enum('pending','active','disabled') NOT NULL DEFAULT 'pending',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `Projects` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `shortDescription` varchar(1024) NOT NULL DEFAULT '',
  `description` text NOT NULL,
  `status` enum('pending','active','disabled') NOT NULL DEFAULT 'pending',
  `showStatus` enum('site','direct','no') NOT NULL DEFAULT 'site',
  `signupStatus` enum('open','with_code','closed') NOT NULL DEFAULT 'open',
  `maxParticipants` int(6) NOT NULL DEFAULT 0,
  `participantVisibility` enum('code','email','full') NOT NULL DEFAULT 'code',
  `participantMinimumAge` int(3) NOT NULL DEFAULT 0,
  `connectParticipantToConsentForm` enum('yes','no') DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `Flows` (
  `projectId` int(11) NOT NULL,
  `itemId` int(11) NOT NULL,
  `itemType` enum('consent','module') NOT NULL DEFAULT 'module',
  `flowOrder` int(11) NOT NULL,
  KEY `projectId` (`projectId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `Modules` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `projectId` int(11) NOT NULL,
  `name` varchar(128) NOT NULL,
  `status` enum('pending','active','disabled') DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `Blocks` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `moduleId` int(11) NOT NULL,
  `moduleOrder` int(11) NOT NULL DEFAULT 0,
  `blockType` enum('other','sign_up','survey','presentation') NOT NULL DEFAULT 'other',
  `blockTypeId` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `moduleId` (`moduleId`),
  KEY `blockType` (`blockType`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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


CREATE TABLE `ConsentForms` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `projectId` int(11) NOT NULL,
  `contentInMarkdown` text NOT NULL,
  `contactInformationDisplay` varchar(512) NOT NULL DEFAULT '',
  `institutionInformationDisplay` varchar(512) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `projectId` (`projectId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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

CREATE TABLE `Tokens` (
  `userId` int(11) NOT NULL,
  `tokenType` enum('email','password_reset','refresh') NOT NULL DEFAULT 'email',
  `createdOn` datetime NOT NULL,
  `expiresOn` datetime NOT NULL,
  `token` varchar(128) NOT NULL,
  UNIQUE KEY `userId` (`userId`,`tokenType`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;