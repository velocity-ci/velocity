defmodule Velocity.UserRepository do
  @moduledoc "provides user queries"

  alias Velocity.User
  alias Velocity.Repo

  def find_by_username(username) do
    user = Repo.get_by(User, username: username)
    if user != nil do
      {:ok, user}
    else
      {:error}
    end
  end
end
