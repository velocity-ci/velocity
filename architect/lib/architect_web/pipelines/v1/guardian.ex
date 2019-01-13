defmodule Architect.Pipelines.Guardian do
  use Guardian.Plug.Pipeline,
    otp_app: :auth_me,
    error_handler: ArchitectWeb.V1.UserController,
    module: Architect.Users.Guardian

  # If there is an authorization header, restrict it to an access token and validate it
  plug(Guardian.Plug.VerifyHeader, claims: %{"typ" => "access"})
  # Load the user if either of the verifications worked
  plug(Guardian.Plug.LoadResource, allow_blank: false)
end
