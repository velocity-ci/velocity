module Page.Project exposing (..)

import Data.Project as Project exposing (Project)
import Data.Commit as Commit exposing (Commit)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.Project
import Task exposing (Task)
import Data.Session as Session exposing (Session)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Views.Page as Page
import Util exposing ((=>))
import Http
import Page.Helpers exposing (formatDate)


-- MODEL --


type alias Model =
    { project : Project
    , commits : List Commit
    }


init : Session -> Project.Id -> Task PageLoadError Model
init session id =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadProject =
            maybeAuthToken
                |> Request.Project.get id
                |> Http.toTask

        loadCommits =
            maybeAuthToken
                |> Request.Project.commits id
                |> Http.toTask

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."
    in
        Task.map2 Model loadProject loadCommits
            |> Task.mapError handleLoadError



-- VIEW --


view : Session -> Model -> Html Msg
view session model =
    let
        project =
            model.project
    in
        div [ class "card" ]
            [ h3 [ class "card-header" ]
                [ text project.name ]
            , div [ class "card-block" ]
                [ div [] [ button [ class "btn btn-primary btn-lg", onClick SubmitSync ] [ text "Synchronize" ] ]
                , div [] (List.map viewCommitListItem model.commits)
                ]
            ]


viewCommitListItem : Commit -> Html Msg
viewCommitListItem commit =
    let
        authorAndDate =
            [ strong [] [ text commit.author ], text (" commited on " ++ formatDate commit.date) ]

        truncatedHash =
            String.slice 0 7 commit.hash
    in
        div [ class "list-group" ]
            [ a [ class "list-group-item list-group-item-action flex-column align-items-start", href "#" ]
                [ div [ class "d-flex w-100 justify-content-between" ]
                    [ h5 [ class "mb-1" ] [ text commit.message ]
                    , small [] [ text truncatedHash ]
                    ]
                , small [] authorAndDate
                ]
            ]



-- UPDATE --


type Msg
    = SubmitSync
    | SyncCompleted (Result Http.Error Project)


update : Session -> Msg -> Model -> ( Model, Cmd Msg )
update session msg model =
    case msg of
        SubmitSync ->
            let
                cmdFromAuth authToken =
                    authToken
                        |> Request.Project.sync model.project.id
                        |> Http.send SyncCompleted

                cmd =
                    session
                        |> Session.attempt "sync project" cmdFromAuth
                        |> Tuple.second
            in
                model => cmd

        SyncCompleted (Ok project) ->
            model => Cmd.none

        SyncCompleted (Err err) ->
            model => Cmd.none
