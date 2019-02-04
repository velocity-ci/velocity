defmodule ArchitectWeb.Schema do
  use Absinthe.Schema
  use Absinthe.Relay.Schema, :modern

  import Kronky.Payload

  alias ArchitectWeb.{Schema, Mutations, Queries, Subscriptions}
  alias Architect.Projects
  alias Architect.Projects.{Project, Task}
  alias Architect.Events
  alias Architect.Events.Event
  alias Git.{Commit, Branch}
  alias ArchitectWeb.{Resolvers, Middleware}
  alias Ecto.Changeset

  # Custom
  import_types(Absinthe.Type.Custom)
  import_types(Kronky.ValidationMessageTypes)

  connection(node_type: :project)
  connection(node_type: :commit)
  connection(node_type: :branch)
  connection(node_type: :event)

  import_types(Schema.UsersTypes)
  import_types(Schema.KnownHostsTypes)
  import_types(Schema.ProjectsTypes)
  import_types(Schema.EventsTypes)

  import_types(Mutations.KnownHostsMutations)
  import_types(Mutations.ProjectsMutations)
  import_types(Mutations.AuthMutations)
  import_types(Queries.KnownHostsQueries)
  import_types(Queries.ProjectsQueries)
  import_types(Subscriptions.KnownHostSubscriptions)
  import_types(Subscriptions.ProjectsSubscriptions)

  payload_object(:session_payload, :session)
  payload_object(:known_host_payload, :known_host)
  payload_object(:project_payload, :project)
  require Logger

  node interface do
    resolve_type(fn
      %Project{}, _ ->
        :project

      %Branch{}, _ ->
        :branch

      %Commit{}, _ ->
        :commit

      %Task{}, _ ->
        :task

      %Event{}, _ ->
        :event

      _, _ ->
        nil
    end)
  end

  mutation do
    import_fields(:auth_mutations)
    import_fields(:known_hosts_mutations)
    import_fields(:projects_mutations)
  end

  query do
    import_fields(:known_hosts_queries)
    #    import_fields(:projects_queries)

    @desc "List projects"
    connection field(:projects, node_type: :project) do
      #      middleware(Middleware.Authorize)
      resolve(&Resolvers.Projects.list_projects/2)
    end

    @desc "List events"
    connection field(:events, node_type: :event) do
      #      middleware(Middleware.Authorize)
      resolve(&Resolvers.Events.list_events/2)
    end

    @desc "List commits"
    connection field(:commits, node_type: :commit) do
      #      middleware(Middleware.Authorize)

      arg(:project_slug, non_null(:string))
      arg(:branch, non_null(:string))

      # Add the project to the context
      middleware(fn res, _ ->
        [%{argument_data: %{project_slug: project_slug}} | _] = res.path
        project = Projects.get_project_by_slug!(project_slug)
        %{res | context: Map.put(res.context, :project, project)}
      end)

      resolve(&Resolvers.Projects.list_commits/2)
    end

    @desc "Get branch"
    field(:branch, non_null(:branch)) do
      #      middleware(Middleware.Authorize)

      arg(:project_slug, non_null(:string))
      arg(:branch, non_null(:string))

      # Add the project to the context
      middleware(fn res, _ ->
        [%{argument_data: %{project_slug: project_slug}} | _] = res.path
        project = Projects.get_project_by_slug!(project_slug)
        %{res | context: Map.put(res.context, :project, project)}
      end)

      resolve(&Resolvers.Projects.get_branch/2)
    end

    @desc "Get project"
    field(:project, non_null(:project)) do
      arg(:slug, non_null(:string))
      resolve(&Resolvers.Projects.get_project/2)
    end
  end

  subscription do
    import_fields(:known_hosts_subscriptions)
    import_fields(:projects_subscriptions)
  end

  def middleware(middleware, _field, %Absinthe.Type.Object{identifier: :mutation}) do
    middleware ++ [&build_payload/2]
  end

  def middleware(middleware, _field, _object) do
    middleware
  end

  def plugins do
    [Absinthe.Middleware.Dataloader | Absinthe.Plugin.defaults()]
  end

  def dataloader() do
    Dataloader.new()
  end

  def context(ctx) do
    Map.put(ctx, :loader, dataloader())
  end
end
