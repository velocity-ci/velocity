defmodule VelocityWeb.AuthView do
  @moduledoc "provides authentication output\n"

  use VelocityWeb, :view

  @spec render(String, {:user, :jwt, :exp}) :: {}
  def render("auth.json", %{user: user, jwt: jwt, exp: exp}) do
    %{data: %{username: user.username, authToken: jwt, exp: exp}}
  end
end
