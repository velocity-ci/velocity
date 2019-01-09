defmodule ArchitectWeb.V1.UserView do
  use ArchitectWeb, :view
  alias ArchitectWeb.V1.UserView

  def render("index.json", %{users: users}) do
    %{data: render_many(users, UserView, "user.json")}
  end

  def render("show.json", %{user: user}) do
    %{data: render_one(user, UserView, "user.json")}
  end

  def render("user.json", %{user: user}) do
    %{id: user.id, username: user.username}
  end

  def render("auth.json", %{user: user, token: token, claims: claims}) do
    %{
      username: user.username,
      token: token,
      claims: claims
    }
  end

  def render("auth_error.json", _) do
    %{
      errors: %{
        detail: "unauthorized"
      }
    }
  end
end
