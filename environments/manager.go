package environments

type EnvManager struct {
	envs map[string]Actions
}

func NewEnvManager() *EnvManager {
	return &EnvManager{
		envs: map[string]Actions{},
	}
}

func (e *EnvManager) RegisterEnv(id string, env Actions) {
	e.envs[id] = env
}

func (e *EnvManager) Env(env string) Actions {
	return e.envs[env]
}
