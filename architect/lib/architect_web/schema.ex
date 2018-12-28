defmodule ArchitectWeb.Schema do
  use Absinthe.Schema

  import Kronky.Payload
  alias ArchitectWeb.Resolvers
  alias ArchitectWeb.Schema.Middleware.TranslateMessages

  import_types(Kronky.ValidationMessageTypes)
  import_types(ArchitectWeb.Schema.UsersTypes)
  import_types(ArchitectWeb.Schema.KnownHostsTypes)
  import_types(ArchitectWeb.Schema.ProjectsTypes)
  import_types(ArchitectWeb.Mutations.UsersMutations)
  import_types(ArchitectWeb.Mutations.AuthMutations)
  import_types(ArchitectWeb.Queries.UsersQueries)
  import_types(ArchitectWeb.Queries.KnownHostsQueries)
  import_types(ArchitectWeb.Queries.ProjectsQueries)

  payload_object(:user_payload, :user)
  payload_object(:session_payload, :session)

  mutation do
    import_fields(:user_mutations)
    import_fields(:auth_mutations)
  end

  query do
    import_fields(:users_queries)
    import_fields(:known_hosts_queries)
    import_fields(:projects_queries)
  end

  #  def middleware(middleware, %Absinthe.Type.Field{identifier: :sign_in}, %Absinthe.Type.Object{
  #        identifier: :mutation
  #      }) do
  #    IO.inspect("ignored", label: "DAT FIELD")
  #
  #    middleware
  #  end
  #
  def middleware(middleware, field, %Absinthe.Type.Object{identifier: :mutation}) do
    IO.inspect("not ignored", label: "DAT FIELD")

    middleware ++ [&build_payload/2, TranslateMessages]
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
