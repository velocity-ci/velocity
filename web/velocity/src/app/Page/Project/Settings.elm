module Page.Project.Settings exposing (..)

import Context exposing (Context)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Util exposing ((=>))
import Route exposing (Route)
import Page.Project.Route as ProjectRoute
import Request.Project
import Request.Errors
import Route
import Task


-- MODEL --


type ConfirmDeleteState
    = Open Bool String
    | Closed


type alias Model =
    ConfirmDeleteState


initialModel : Model
initialModel =
    Closed



-- VIEW --


view : Project -> Model -> Html Msg
view project model =
    div []
        [ viewDangerArea project model ]


viewDangerArea : Project -> Model -> Html Msg
viewDangerArea project model =
    let
        cardBody =
            case model of
                Closed ->
                    viewPreDeleteConfirmation model

                Open deleting deleteConfirmProjectName ->
                    viewDeleteConfirmation project.name deleteConfirmProjectName deleting
    in
        div [ class "card border-danger mb-3" ]
            [ div [ class "card-body" ]
                [ h5 [ class "card-title" ] [ text "Delete project" ]
                , cardBody
                ]
            ]


viewPreDeleteConfirmation : Model -> Html Msg
viewPreDeleteConfirmation model =
    div []
        [ p
            [ class "card-text" ]
            [ text "Once you delete a project, there is no going back. Please be certain." ]
        , button
            [ type_ "button"
            , class "btn btn-outline-danger"
            , onClick (SetDeleteState (Open False ""))
            ]
            [ text "Delete project" ]
        ]


viewDeleteConfirmation : String -> String -> Bool -> Html Msg
viewDeleteConfirmation projectName confirmValue submitting =
    let
        disclaimer =
            div []
                [ p []
                    [ text "This will permanently delete the "
                    , strong [] [ text projectName ]
                    , text " project and builds."
                    ]
                , p []
                    [ text "Please type in the name of the project to confirm or "
                    , button
                        [ type_ "button"
                        , class "btn btn-link btn-cancel-delete"
                        , onClick (SetDeleteState Closed)
                        , disabled submitting
                        ]
                        [ text "click here to cancel." ]
                    ]
                ]
    in
        div []
            [ disclaimer
            , div [ class "input-group" ]
                [ input
                    [ class "form-control"
                    , type_ "text"
                    , value confirmValue
                    , onInput ((Open False) >> SetDeleteState)
                    , disabled submitting
                    ]
                    []
                , span [ class "input-group-btn" ]
                    [ button
                        [ class "btn btn-danger"
                        , type_ "button"
                        , disabled ((projectName /= confirmValue) || submitting)
                        , onClick SubmitProjectDelete
                        ]
                        [ text "Delete project" ]
                    ]
                ]
            ]


breadcrumb : Project -> List ( Route, String )
breadcrumb project =
    [ ( Route.Project project.slug ProjectRoute.Settings, "Settings" ) ]



-- UPDATE --


type Msg
    = SubmitProjectDelete
    | ProjectDeleted (Result Request.Errors.HttpError ())
    | SetDeleteState ConfirmDeleteState


update : Context -> Project -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update context project session msg model =
    case msg of
        SubmitProjectDelete ->
            let
                cmdFromAuth authToken =
                    authToken
                        |> Request.Project.delete context project.slug
                        |> Task.attempt ProjectDeleted

                cmd =
                    session
                        |> Session.attempt "delete project" cmdFromAuth
                        |> Tuple.second
            in
                case model of
                    Open _ value ->
                        Open True value => cmd

                    Closed ->
                        Open True "" => cmd

        ProjectDeleted (Ok _) ->
            Open True "" => Route.modifyUrl Route.Projects

        ProjectDeleted (Err _) ->
            Open False "" => Cmd.none

        SetDeleteState state ->
            state => Cmd.none
