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
import Route
import Util exposing (onClickStopPropagation)


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
        div []
            [ viewBreadcrumb project
            , div [ class "container-fluid" ]
                [ div [ class "row" ]
                    [ viewSidebar
                    , viewProjectContainer model
                    ]
                ]
            ]


viewBreadcrumb : Project -> Html Msg
viewBreadcrumb project =
    div [ class "d-flex justify-content-start align-items-center bg-dark", style [ ( "height", "50px" ) ] ]
        [ div [ class "p-2" ]
            [ ol [ class "breadcrumb bg-dark", style [ ( "margin", "0" ) ] ]
                [ li [ class "breadcrumb-item" ] [ a [ Route.href Route.Projects ] [ text "Projects" ] ]
                , li [ class "breadcrumb-item active" ] [ text project.name ]
                ]
            ]
        , div [ class "ml-auto p-2" ]
            [ button [ class "ml-auto btn btn-primary", type_ "button", onClick SubmitSync ] [ text "Synchronize" ]
            ]
        ]


viewProjectContainer : Model -> Html Msg
viewProjectContainer model =
    div [ class "col-sm-9 ml-sm-auto col-md-10 pt-3" ]
        [ viewCommitList model.commits
        ]


viewCommitList : List Commit -> Html Msg
viewCommitList commits =
    List.map viewCommitListItem commits
        |> div []


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


viewSidebar : Html Msg
viewSidebar =
    nav [ class "col-sm-3 col-md-2 d-none d-sm-block bg-light sidebar" ]
        [ ul [ class "nav nav-pills flex-column" ]
            [ li [ class "nav-item" ]
                [ a [ class "nav-link active", href "#" ]
                    [ text "Commits "
                    , span [ class "sr-only" ] [ text "(current)" ]
                    ]
                ]
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
