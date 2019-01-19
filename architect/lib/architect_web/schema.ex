defmodule ArchitectWeb.Schema do
  use Absinthe.Schema

  import Kronky.Payload

  alias ArchitectWeb.{Schema, Mutations, Queries, Subscriptions}

  # Custom
  import_types(Absinthe.Type.Custom)
  import_types(Kronky.ValidationMessageTypes)

  import_types(Schema.UsersTypes)
  import_types(Schema.KnownHostsTypes)
  import_types(Schema.ProjectsTypes)
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

  mutation do
    import_fields(:auth_mutations)
    import_fields(:known_hosts_mutations)
    import_fields(:projects_mutations)
  end

  query do
    import_fields(:known_hosts_queries)
    import_fields(:projects_queries)
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