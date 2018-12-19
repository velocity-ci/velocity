defmodule ArchitectWeb.Schema.UsersTypes do
  use Absinthe.Schema.Notation

  object :user do
    field(:id, :id)
    field(:username, :string)
    field(:password, :string)
  end

  @desc "token to authenticate user"
  object :session do
    field(:token, :string)
  end
end
