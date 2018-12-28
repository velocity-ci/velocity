defmodule ArchitectWeb.Queries.UsersQueries do
  use Absinthe.Schema.Notation
  alias ArchitectWeb.Resolvers

  object :users_queries do
    @desc "Get all users"
    field :users, list_of(:user) do
      resolve(&Resolvers.Users.list_users/3)
    end
  end
end
