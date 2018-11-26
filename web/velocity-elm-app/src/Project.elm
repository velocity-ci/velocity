module Project exposing (Project, addProject, channel, channelName, create, decoder, id, list, name, repository, slug, sync, syncing, thumbnail, thumbnailSrc, updateProject)

import Api exposing (BaseUrl, Cred)
import Api.Endpoint as Endpoint exposing (Endpoint)
import Element exposing (..)
import Element.Background as Background
import Element.Border as Border
import Element.Font as Font
import Http
import Icon
import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import Json.Encode as Encode
import PaginatedList exposing (PaginatedList)
import Palette
import Phoenix.Channel as Channel exposing (Channel)
import Project.Id as Id exposing (Id)
import Project.Slug as Slug exposing (Slug)
import Time


type Project
    = Project Internals


type alias Internals =
    { id : Id
    , slug : Slug
    , name : String
    , repository : String
    , createdAt : Time.Posix
    , updatedAt : Time.Posix
    , synchronising : Bool
    , logo : Maybe String
    }



-- SERIALIZATION --


decoder : Decoder Project
decoder =
    Decode.succeed Project
        |> custom internalsDecoder


internalsDecoder : Decoder Internals
internalsDecoder =
    Decode.succeed Internals
        |> required "id" Id.decoder
        |> required "slug" Slug.decoder
        |> required "name" Decode.string
        |> required "repository" Decode.string
        |> required "createdAt" Iso8601.decoder
        |> required "updatedAt" Iso8601.decoder
        |> required "synchronising" Decode.bool
        |> required "logo" (Decode.maybe Decode.string)



-- INFO --


id : Project -> Id
id (Project project) =
    project.id


name : Project -> String
name (Project project) =
    project.name


slug : Project -> Slug
slug (Project project) =
    project.slug


thumbnailSrc : Project -> Maybe String
thumbnailSrc (Project project) =
    project.logo


repository : Project -> String
repository (Project project) =
    project.repository


channelName : Project -> String
channelName (Project project) =
    "project:" ++ Slug.toString project.slug


channel : Project -> Channel msg
channel project =
    Channel.init (channelName project)


syncing : Project -> Bool
syncing (Project project) =
    project.synchronising



-- ELEMENTS --


thumbnail : Project -> Element msg
thumbnail project =
    case thumbnailSrc project of
        Just src ->
            el
                [ width (px 100)
                , height (px 100)
                , Background.uncropped src
                , Border.width 1
                , Border.color Palette.neutral5
                , Border.rounded 10
                , padding 5
                ]
                (text "")

        Nothing ->
            el
                [ width (px 100)
                , height (px 100)
                , Border.width 1
                , Border.color Palette.neutral5
                , Border.rounded 10
                , paddingXY 5 0
                , Font.color Palette.neutral6
                ]
                (Icon.code Icon.fullSizeOptions)



-- HELPERS --


findProject : List Project -> Project -> Maybe Project
findProject projects (Project a) =
    List.filter (\(Project b) -> b.id == a.id) projects
        |> List.head


addProject : Project -> List Project -> List Project
addProject project projects =
    case findProject projects project of
        Just _ ->
            projects

        Nothing ->
            project :: projects


updateProject : Project -> List Project -> List Project
updateProject (Project a) projects =
    projects
        |> List.map
            (\(Project b) ->
                if a.id == b.id then
                    Project a

                else
                    Project b
            )



-- COLLECTION --


list : Cred -> BaseUrl -> Http.Request (List Project)
list cred baseUrl =
    let
        endpoint =
            Endpoint.projects (Just { amount = -1, page = 1 }) (Api.toEndpoint baseUrl)
    in
    Decode.field "data" (Decode.list decoder)
        |> Api.get endpoint (Just cred)


sync : Cred -> BaseUrl -> Slug -> Http.Request Project
sync cred baseUrl slug_ =
    let
        endpoint =
            Endpoint.projectSync (Api.toEndpoint baseUrl) slug_
    in
    Api.post endpoint (Just cred) (Encode.object [] |> Http.jsonBody) decoder


create : Cred -> BaseUrl -> { a | name : String, repository : String, privateKey : Maybe String } -> Http.Request Project
create cred baseUrl values =
    let
        endpoint =
            Endpoint.projects Nothing (Api.toEndpoint baseUrl)

        baseValues =
            [ ( "name", Encode.string values.name )
            , ( "address", Encode.string values.repository )
            ]

        submitValues =
            case values.privateKey of
                Just privateKey ->
                    ( "key", Encode.string privateKey ) :: baseValues

                Nothing ->
                    baseValues

        body =
            submitValues
                |> Encode.object
                |> Http.jsonBody
    in
    Api.post endpoint (Just cred) body decoder



-- CHANNEL --
