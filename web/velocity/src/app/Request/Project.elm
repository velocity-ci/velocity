module Request.Project exposing (list, create, get)

import Data.AuthToken as AuthToken exposing (AuthToken, withAuthorization)
import Data.Project as Project exposing (Project)
import Json.Decode as Decode
import Json.Encode as Encode
import Request.Helpers exposing (apiUrl)
import HttpBuilder exposing (RequestBuilder, withBody, withExpect, withQueryParams)
import Util exposing ((=>))
import Http


baseUrl : String
baseUrl =
    "/projects"



-- LIST --


list : Maybe AuthToken -> Http.Request (List Project)
list maybeToken =
    let
        expect =
            Project.decoder
                |> Decode.list
                |> Http.expectJson
    in
        apiUrl baseUrl
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest



-- GET --


get : Project.Id -> Maybe AuthToken -> Http.Request Project
get id maybeToken =
    let
        expect =
            Project.decoder
                |> Http.expectJson
    in
        apiUrl (baseUrl ++ "/" ++ Project.idToString id)
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest



-- CREATE --


type alias CreateConfig record =
    { record
        | name : String
        , repository : String
        , privateKey : String
    }


create : CreateConfig record -> AuthToken -> Http.Request Project
create config token =
    let
        expect =
            Project.decoder
                |> Http.expectJson

        project =
            Encode.object
                [ "name" => Encode.string config.name
                , "repository" => Encode.string config.repository
                , "key" => Encode.string config.privateKey
                ]

        body =
            project
                |> Http.jsonBody
    in
        apiUrl baseUrl
            |> HttpBuilder.post
            |> withAuthorization (Just token)
            |> withBody body
            |> withExpect expect
            |> HttpBuilder.toRequest
