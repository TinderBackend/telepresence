//go:generate ./generic.gen InterceptMap *github.com/TinderBackend/telepresence/rpc/v2/manager.InterceptInfo Id
//go:generate ./generic.gen AgentMap     *github.com/TinderBackend/telepresence/rpc/v2/manager.AgentInfo     Name
//go:generate ./generic.gen ClientMap    *github.com/TinderBackend/telepresence/rpc/v2/manager.ClientInfo    Name

package watchable
