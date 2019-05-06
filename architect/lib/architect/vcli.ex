defmodule Architect.VCLI do
  @moduledoc """
  Deals with calling the vcli binary
  """

  require Logger

  def init(), do: Application.get_env(:architect, :vcli) |> Enum.into(%{})

  def list(dir, opts), do: cmd(dir, opts, ["list", "--machine-readable"])

  def project_config(dir, opts), do: cmd(dir, opts, ["info", "--machine-readable"])

  def plan_blueprint(dir, opts, task_name), do: cmd(dir, opts, ["run", task_name, "--plan-only", "--machine-readable"])

  defp cmd(dir, %{bin: bin, timeout: timeout, log_errors: log_errors}, cmd) when is_list(cmd) do
    try do
      %Porcelain.Result{err: nil, out: out, status: 0} =
        Task.async(fn ->
          Porcelain.exec(bin, cmd, dir: dir)
        end)
        |> Task.await(timeout)

      Poison.decode!(out)
    catch
      :exit, _ ->
        log("VCLI timeout", :error, log_errors)
        {:error, :timeout}
    end
  end

  @doc false
  defp log(output, :debug, true), do: Logger.debug(output)
  defp log(output, :warn, true), do: Logger.warn(output)
  defp log(output, :error, true), do: Logger.error(output)
  defp log(_, _, _), do: nil
end
