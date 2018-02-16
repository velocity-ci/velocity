module Request.Project
    exposing
        ( list
        , create
        , get
        , sync
        , delete
        , branches
        , builds
        )

import Data.AuthToken as AuthToken exposing (AuthToken, withAuthorization)
import Data.Project as Project exposing (Project)
import Data.Branch as Branch exposing (Branch)
import Data.PaginatedList as PaginatedList exposing (PaginatedList)
import Data.Build as Build exposing (Build)
import Json.Encode as Encode
import Request.Helpers exposing (apiUrl)
import HttpBuilder exposing (RequestBuilder, withBody, withExpect, withQueryParams)
import Util exposing ((=>))
import Http


baseUrl : String
baseUrl =
    "/projects"



-- LIST --


list : Maybe AuthToken -> Http.Request (PaginatedList Project)
list maybeToken =
    let
        expect =
            Project.decoder
                |> PaginatedList.decoder
                |> Http.expectJson
    in
        apiUrl baseUrl
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest



-- SYNC --


sync : Project.Id -> AuthToken -> Http.Request Project
sync id token =
    let
        expect =
            Project.decoder
                |> Http.expectJson
    in
        apiUrl (baseUrl ++ "/" ++ Project.idToString id ++ "/sync")
            |> HttpBuilder.post
            |> withAuthorization (Just token)
            |> withExpect expect
            |> HttpBuilder.toRequest



-- BRANCHES --


branches : Project.Id -> Maybe AuthToken -> Http.Request (PaginatedList Branch)
branches id maybeToken =
    let
        expect =
            Branch.decoder
                |> PaginatedList.decoder
                |> Http.expectJson
    in
        apiUrl (baseUrl ++ "/" ++ Project.idToString id ++ "/branches")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest



-- BUILDS --


builds : Project.Id -> Maybe AuthToken -> Http.Request (PaginatedList Build)
builds id maybeToken =
    let
        expect =
            Build.decoder
                |> PaginatedList.decoder
                |> Http.expectJson
    in
        apiUrl (baseUrl ++ "/" ++ Project.idToString id ++ "/builds")
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
        , privateKey : Maybe String
    }


create : CreateConfig record -> AuthToken -> Http.Request Project
create config token =
    let
        expect =
            Project.decoder
                |> Http.expectJson

        baseProject =
            [ "name" => Encode.string config.name
            , "address" => Encode.string config.repository
            ]

        project =
            config.privateKey
                |> Maybe.map (\privateKey -> ( "key", Encode.string privateKey ) :: baseProject)
                |> Maybe.withDefault baseProject

        body =
            project
                |> Encode.object
                |> Http.jsonBody
    in
        apiUrl baseUrl
            |> HttpBuilder.post
            |> withAuthorization (Just token)
            |> withBody body
            |> withExpect expect
            |> HttpBuilder.toRequest



-- DELETE --


delete : Project.Id -> AuthToken -> Http.Request ()
delete id token =
    apiUrl (baseUrl ++ "/" ++ Project.idToString id)
        |> HttpBuilder.delete
        |> withAuthorization (Just token)
        |> withExpect (Http.expectStringResponse (\_ -> Ok ()))
        |> HttpBuilder.toRequest
