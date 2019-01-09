defmodule ArchitectWeb.Router do
  use ArchitectWeb, :router

  scope "/" do
    forward "/v1", ArchitectWeb.V1Router
    forward "/v2", ArchitectWeb.V2Router
  end
end
