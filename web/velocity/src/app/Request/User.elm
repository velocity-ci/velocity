module Request.User exposing (login, storeSession)

import Context exposing (Context)
import Data.User as User exposing (User)
import Http
import Json.Encode as Encode
import Ports
import Request.Helpers exposing (apiUrl)
import Request.Errors
import Util exposing ((=>))
import Task exposing (Task)


storeSession : User -> Cmd msg
storeSession user =
    User.encode user
        |> Encode.encode 0
        |> Just
        |> Ports.storeSession


login : Context -> { r | username : String, password : String } -> Task Request.Errors.HttpError User
login context { username, password } =
    let
        user =
            Encode.object
                [ "username" => Encode.string username
                , "password" => Encode.string password
                ]

        body =
            user |> Http.jsonBody
    in
        User.decoder
            |> Http.post (apiUrl context "/auth") body
            |> Http.toTask
            |> Task.mapError Request.Errors.handleHttpError
