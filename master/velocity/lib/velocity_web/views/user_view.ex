defmodule VelocityWeb.UserView do
  @moduledoc "provides user output\n"

  use VelocityWeb, :view

  @spec render(String, {}) :: {}
  def render("index.json", %{users: users}) do
    %{data: render_many(users, VelocityWeb.UserView, "user.json")}
  end

  @spec render(String, {}) :: {}
  def render("show.json", %{user: user}) do
    %{data: render_one(user, VelocityWeb.UserView, "user.json")}
  end

  @spec render(String, {:user}) :: {}
  def render("user.json", %{user: user}) do
    %{data: %{username: user.username}}
  end
end
