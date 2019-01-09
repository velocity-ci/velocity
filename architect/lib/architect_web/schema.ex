defmodule ArchitectWeb.Schema do
  use Absinthe.Schema

  import Kronky.Payload
  alias Ecto.Changeset
  alias Architect.KnownHosts.KnownHost
  alias ArchitectWeb.Schema.Middleware.TranslateMessages
  alias ArchitectWeb.{Schema, Mutations, Queries}

  # Custom
  import_types(Absinthe.Type.Custom)
  import_types(Kronky.ValidationMessageTypes)

  import_types(Schema.UsersTypes)
  import_types(Schema.KnownHostsTypes)
  import_types(Schema.ProjectsTypes)
  import_types(Mutations.UsersMutations)
  import_types(Mutations.KnownHostsMutations)
  import_types(Mutations.ProjectsMutations)
  import_types(Mutations.AuthMutations)
  import_types(Queries.UsersQueries)
  import_types(Queries.KnownHostsQueries)
  import_types(Queries.ProjectsQueries)

  payload_object(:user_payload, :user)
  payload_object(:session_payload, :session)
  payload_object(:known_host_payload, :known_host)
  payload_object(:project_payload, :project)

  mutation do
    import_fields(:user_mutations)
    import_fields(:auth_mutations)
    import_fields(:known_hosts_mutations)
    import_fields(:projects_mutations)
  end

  query do
    import_fields(:users_queries)
    import_fields(:known_hosts_queries)
    import_fields(:projects_queries)
  end

  subscription do
    field :known_host_added, non_null(:known_host) do
      trigger(:for_host,
        topic: fn
          %KnownHost{id: id} ->
            ["all", id]

          %Changeset{} ->
            []
        end
      )

      config(fn args, _info ->
        {:ok, topic: "all"}
      end)
    end

    field :known_host_verified, non_null(:known_host) do
      trigger(:verify,
        topic: fn
          %KnownHost{id: id} ->
            ["all", id]

          %Changeset{} ->
            []
        end
      )

      config(fn args, _info ->
        {:ok, topic: "all"}
      end)
    end
  end

  #  def middleware(middleware, %Absinthe.Type.Field{identifier: :sign_in}, %Absinthe.Type.Object{
  #        identifier: :mutation
  #      }) do
  #    IO.inspect("ignored", label: "DAT FIELD")
  #
  #    middleware
  #  end

  def middleware(middleware, _field, %Absinthe.Type.Object{identifier: :mutation}) do
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
