defmodule ArchitectWeb.Schema.UsersTypes do
  use Absinthe.Schema.Notation

  object :user do
    field(:id, :id)
    field(:username, :string)
    field(:password, :string)
  end
end
