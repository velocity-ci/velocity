module Request.User exposing (login, create, list, storeSession)

import Context exposing (Context)
import Data.AuthToken as AuthToken exposing (AuthToken, withAuthorization)
import Data.User as User exposing (User)
import Http
import HttpBuilder exposing (RequestBuilder, withBody, withExpect, withQueryParams)
import Json.Encode as Encode
import Json.Decode as Decode
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


list : Context -> Maybe AuthToken -> Task Request.Errors.HttpError (List User.Username)
list context maybeAuthToken =
    let
        expect =
            User.usernameDecoder
                |> Decode.at [ "username" ]
                |> Decode.list
                |> Decode.at [ "data" ]
                |> Http.expectJson
    in
        apiUrl context "/users?amount=-1"
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeAuthToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleHttpError


create :
    Context
    -> Maybe AuthToken
    -> { a | username : String, password : String }
    -> Task Request.Errors.HttpError User.Username
create context maybeAuthToken { username, password } =
    let
        encoded =
            [ "username" => Encode.string username
            , "password" => Encode.string password
            ]

        body =
            encoded
                |> Encode.object
                |> Http.jsonBody

        expect =
            User.usernameDecoder
                |> Decode.at [ "username" ]
                |> Http.expectJson
    in
        apiUrl context "/users"
            |> HttpBuilder.post
            |> HttpBuilder.withExpect expect
            |> withBody body
            |> withAuthorization maybeAuthToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleHttpError


delete :
    Context
    -> Maybe AuthToken
    -> User.Username
    -> Task Request.Errors.HttpError User.Username
delete context maybeAuthToken username =
    let
        usernameString =
            User.usernameToString username

        expect =
            User.usernameDecoder
                |> Decode.at [ "username" ]
                |> Http.expectJson
    in
        apiUrl context ("/users/" ++ usernameString)
            |> HttpBuilder.delete
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeAuthToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleHttpError
