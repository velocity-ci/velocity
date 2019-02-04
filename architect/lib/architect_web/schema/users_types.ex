defmodule ArchitectWeb.Schema.UsersTypes do
  use Absinthe.Schema.Notation

  @desc "token to authenticate user"
  object :session do
    field(:token, non_null(:string))
    field(:username, non_null(:string))
  end

  object :user do
    field(:username, non_null(:string))
  end
end
