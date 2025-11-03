package backups

import (
	"postgresus-backend/internal/features/backups/backups/usecases"
	backups_config "postgresus-backend/internal/features/backups/config"
	"postgresus-backend/internal/features/databases"
	"postgresus-backend/internal/features/notifiers"
	"postgresus-backend/internal/features/storages"
	"postgresus-backend/internal/features/users"
	user_repositories "postgresus-backend/internal/features/users/repositories"
	"postgresus-backend/internal/util/logger"
	"time"
)

var backupRepository = &BackupRepository{}
var backupService = &BackupService{
	databases.GetDatabaseService(),
	storages.GetStorageService(),
	backupRepository,
	notifiers.GetNotifierService(),
	notifiers.GetNotifierService(),
	backups_config.GetBackupConfigService(),
	usecases.GetCreateBackupUsecase(),
	logger.GetLogger(),
	[]BackupRemoveListener{},
}

var backupBackgroundService = &BackupBackgroundService{
	backupService:       backupService,
	backupRepository:    backupRepository,
	backupConfigService: backups_config.GetBackupConfigService(),
	storageService:      storages.GetStorageService(),
	userService:         users.GetUserService(),
	userRepository:      &user_repositories.UserRepository{},
	databaseService:     databases.GetDatabaseService(),
	lastBackupTime:      time.Now().UTC(),
	logger:              logger.GetLogger(),
}

var backupController = &BackupController{
	backupService,
	users.GetUserService(),
}

func SetupDependencies() {
	backups_config.
		GetBackupConfigService().
		SetDatabaseStorageChangeListener(backupService)

	databases.GetDatabaseService().AddDbRemoveListener(backupService)
	databases.GetDatabaseService().AddDbCopyListener(backups_config.GetBackupConfigService())
}

func GetBackupService() *BackupService {
	return backupService
}

func GetBackupController() *BackupController {
	return backupController
}

func GetBackupBackgroundService() *BackupBackgroundService {
	return backupBackgroundService
}
