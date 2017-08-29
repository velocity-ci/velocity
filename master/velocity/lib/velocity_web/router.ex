defmodule VelocityWeb.Router do
  @moduledoc "provides the routes for the application"
  use VelocityWeb, :router

  pipeline :browser do
    plug :accepts, ["html"]
    plug :fetch_session
    plug :fetch_flash
    plug :protect_from_forgery
    plug :put_secure_browser_headers
  end
  pipeline :api do
    plug :accepts, ["json"]
    plug Guardian.Plug.VerifyHeader, realm: "Bearer"
    plug Guardian.Plug.LoadResource
  end
  scope "/", VelocityWeb do
    pipe_through :browser
    # Use the default browser stack
    get "/", PageController, :index
  end
  # Other scopes may use custom stacks.
  # scope "/api", VelocityWeb do
  #   pipe_through :api
  # end
  scope "/api/v1", VelocityWeb do
    pipe_through :api
    resources "/users", UserController, only: [:create]
    resources "/projects", ProjectController, only: [:create]
    resources "/auth", AuthController, only: [:create]
  end
end
