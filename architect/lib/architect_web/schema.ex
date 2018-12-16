defmodule ArchitectWeb.Schema do
  use Absinthe.Schema

  import Kronky.Payload
  alias ArchitectWeb.Resolvers

  import_types(Kronky.ValidationMessageTypes)
  import_types(ArchitectWeb.Schema.UsersTypes)
  import_types(ArchitectWeb.Schema.KnownHostsTypes)
  import_types(ArchitectWeb.Mutations.UsersMutations)

  payload_object(:user_payload, :user)

  mutation do
    import_fields(:users_migrations)
  end

  query do
    @desc "Get all users"
    field :users, list_of(:user) do
      resolve(&Resolvers.Users.list_users/3)
    end

    @desc "Get all known hosts"
    field :known_hosts, list_of(:known_host) do
      resolve(&Resolvers.KnownHosts.list_known_hosts/3)
    end
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
