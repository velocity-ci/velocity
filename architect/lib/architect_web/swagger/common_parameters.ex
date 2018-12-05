defmodule ArchitectWeb.CommonParameters do
  @moduledoc "Common parameter declarations for phoenix swagger"

  alias PhoenixSwagger.Path.PathObject
  import PhoenixSwagger.Path

  def authorization(path = %PathObject{}) do
    path |> parameter("Authorization", :header, :string, "Bearer access token", required: true)
  end

  def sorting(path = %PathObject{}) do
    path
    |> parameter(:sort_by, :query, :string, "The property to sort by")
    |> parameter(:sort_direction, :query, :string, "The sort direction",
      enum: [:asc, :desc],
      default: :asc
    )
  end
end
