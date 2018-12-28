defmodule ArchitectWeb.Schema.UsersTypes do
  use Absinthe.Schema.Notation

  object :user do
    field(:id, :id)
    field(:username, non_null(:string))
  end

  @desc "token to authenticate user"
  object :session do
    field(:token, non_null(:string))
    field(:username, non_null(:string))
  end
end
