defmodule ArchitectWeb.Schema do
  use Absinthe.Schema
  import_types(ArchitectWeb.Schema.UsersTypes)

  alias ArchitectWeb.Resolvers

  query do
    @desc "Get all users"
    field :users, list_of(:user) do
      resolve(&Resolvers.Users.list_users/3)
    end
  end
end
