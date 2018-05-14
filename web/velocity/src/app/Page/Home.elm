module Page.Home exposing (view, update, Model, Msg, ExternalMsg(..), init, channelName, initialEvents, subscriptions)

{-| The homepage. You can get here via either the / or /#/ routes.
-}

-- EXTERNAL

import Http
import Html exposing (..)
import Html.Attributes exposing (class, href, id, placeholder, attribute, classList, style)
import Html.Events exposing (onClick)
import Task exposing (Task)
import Navigation exposing (newUrl)
import Dict exposing (Dict)
import Time.DateTime as DateTime
import Json.Encode as Encode
import Json.Decode as Decode
import Bootstrap.Modal as Modal
import Bootstrap.Button as Button
import Json.Decode as Decode exposing (decodeString)


-- INTERNAL

import Context exposing (Context)
import Component.ProjectForm as ProjectForm
import Component.Form as Form
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Data.PaginatedList as PaginatedList exposing (Paginated(..))
import Page.Helpers exposing (formatDate, sortByDatetime)
import Page.Project.Route as ProjectRoute
import Views.Helpers exposing (onClickPage)
import Views.Page as Page
import Util exposing ((=>), onClickStopPropagation, viewIf)
import Request.Project
import Request.Errors
import Route


-- MODEL --


type alias Model =
    { projects : List Project
    , newProjectForm : ProjectForm.Context
    , newProjectModalVisibility : Modal.Visibility
    }


init : Context -> Session msg -> Task (Request.Errors.Error PageLoadError) Model
init context session =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadProjects =
            Request.Project.list context maybeAuthToken

        errorPage =
            pageLoadError Page.Home "Homepage is currently unavailable."

        initialModel projects =
            { projects = projects
            , newProjectForm = ProjectForm.init
            , newProjectModalVisibility = Modal.hidden
            }
    in
        Task.map (\(Paginated { results }) -> initialModel results) loadProjects
            |> Task.mapError (Request.Errors.withDefaultError errorPage)



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions { newProjectModalVisibility } =
    Modal.subscriptions newProjectModalVisibility AnimateNewProjectModal



-- CHANNELS --


channelName : String
channelName =
    "projects"


initialEvents : Dict String (List ( String, Encode.Value -> Msg ))
initialEvents =
    let
        pageEvents =
            [ ( "project:new", AddProject ) ]
    in
        Dict.singleton channelName pageEvents



-- VIEW --


view : Session msg -> Model -> Html Msg
view session model =
    let
        hasProjects =
            not (List.isEmpty model.projects)

        projectList =
            viewProjectList model.projects
    in
        div [ class "py-2 my-4" ]
            [ viewToolbar
            , viewIf hasProjects projectList
            , viewNewProjectModal model.newProjectForm model.newProjectModalVisibility
            ]


viewToolbar : Html Msg
viewToolbar =
    div [ class "btn-toolbar d-flex flex-row-reverse" ]
        [ button
            [ class "btn btn-primary btn-lg"
            , style [ "border-radius" => "25px" ]
            , onClick ShowNewProjectModal
            ]
            [ i [ class "fa fa-plus" ] [] ]
        ]


viewProjectList : List Project -> Html Msg
viewProjectList projects =
    let
        latestProjects =
            sortByDatetime .updatedAt projects
    in
        div []
            [ h6 [] [ text "Projects" ]
            , ul [ class "list-group list-group-flush" ] (List.map viewProjectListItem latestProjects)
            ]


projectFormConfig : ProjectForm.Config Msg
projectFormConfig =
    { setNameMsg = SetProjectFormName
    , setRepositoryMsg = SetProjectFormRepository
    , setPrivateKeyMsg = SetProjectFormPrivateKey
    , submitMsg = SubmitProjectForm
    }


viewNewProjectModal : ProjectForm.Context -> Modal.Visibility -> Html Msg
viewNewProjectModal projectForm visibility =
    Modal.config CloseNewProjectModal
        |> Modal.withAnimation AnimateNewProjectModal
        |> Modal.large
        |> Modal.hideOnBackdropClick True
        |> Modal.h3 [] [ text "Create project" ]
        |> Modal.body [] [ ProjectForm.view projectFormConfig projectForm ]
        |> Modal.footer [] [ ProjectForm.viewSubmitButton projectFormConfig projectForm ]
        |> Modal.view visibility


viewProjectListItem : Project -> Html Msg
viewProjectListItem project =
    let
        route =
            Route.Project project.slug ProjectRoute.Overview

        lastUpdatedText =
            "Last updated " ++ formatDate (DateTime.date project.updatedAt)

        smallText =
            project.repository

        projectLink =
            a
                [ Route.href route
                , onClickPage NewUrl route
                ]
                [ text project.name ]
    in
        li [ class "list-group-item flex-column align-items-start px-0" ]
            [ div [ class "d-flex w-100 justify-content-between" ]
                [ h5 [ class "mb-1" ] [ projectLink ]
                , small [] [ text lastUpdatedText ]
                ]
            , small [] [ text smallText ]
            ]



-- UPDATE --


type Msg
    = NewUrl String
    | AddProject Encode.Value
    | CloseNewProjectModal
    | AnimateNewProjectModal Modal.Visibility
    | ShowNewProjectModal
    | SubmitProjectForm
    | SetProjectFormName String
    | SetProjectFormRepository String
    | SetProjectFormPrivateKey String
    | ProjectCreated (Result Request.Errors.HttpError Project)


type ExternalMsg
    = NoOp
    | HandleRequestError Request.Errors.HandledError


findProject : List Project -> Project -> Maybe Project
findProject projects project =
    List.filter (\a -> a.id == project.id) projects
        |> List.head


addProject : List Project -> Project -> List Project
addProject projects project =
    case findProject projects project of
        Just _ ->
            projects

        Nothing ->
            project :: projects


update : Context -> Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update context session msg model =
    case msg of
        NewUrl url ->
            model
                => newUrl url
                => NoOp

        CloseNewProjectModal ->
            { model
                | newProjectModalVisibility = Modal.hidden
                , newProjectForm = ProjectForm.init
            }
                => Cmd.none
                => NoOp

        AnimateNewProjectModal visibility ->
            { model | newProjectModalVisibility = visibility }
                => Cmd.none
                => NoOp

        SetProjectFormName name ->
            { model | newProjectForm = ProjectForm.update model.newProjectForm ProjectForm.Name name }
                => Cmd.none
                => NoOp

        SetProjectFormRepository repository ->
            { model | newProjectForm = ProjectForm.update model.newProjectForm ProjectForm.Repository repository }
                => Cmd.none
                => NoOp

        SetProjectFormPrivateKey privateKey ->
            { model | newProjectForm = ProjectForm.update model.newProjectForm ProjectForm.PrivateKey privateKey }
                => Cmd.none
                => NoOp

        SubmitProjectForm ->
            let
                cmdFromAuth authToken =
                    authToken
                        |> Request.Project.create context (ProjectForm.submitValues model.newProjectForm)
                        |> Task.attempt ProjectCreated

                cmd =
                    session
                        |> Session.attempt "create project" cmdFromAuth
                        |> Tuple.second
            in
                { model | newProjectForm = Form.submit model.newProjectForm }
                    => cmd
                    => NoOp

        ProjectCreated (Err err) ->
            let
                ( updatedProjectForm, externalMsg ) =
                    case err of
                        Request.Errors.HandledError handledError ->
                            model.newProjectForm
                                => HandleRequestError handledError

                        Request.Errors.UnhandledError (Http.BadStatus response) ->
                            let
                                errors =
                                    response.body
                                        |> decodeString ProjectForm.errorsDecoder
                                        |> Result.withDefault []
                            in
                                model.newProjectForm
                                    |> Form.updateServerErrors errors ProjectForm.serverErrorToFormError
                                    => NoOp

                        _ ->
                            model.newProjectForm
                                |> Form.updateServerErrors [ "" => "Unable to process project." ] ProjectForm.serverErrorToFormError
                                => NoOp
            in
                { model | newProjectForm = Form.submitting False updatedProjectForm }
                    => Cmd.none
                    => externalMsg

        ProjectCreated (Ok project) ->
            { model
                | newProjectForm = ProjectForm.init
                , newProjectModalVisibility = Modal.hidden
                , projects = addProject model.projects project
            }
                => Cmd.none
                => NoOp

        ShowNewProjectModal ->
            { model | newProjectModalVisibility = Modal.shown }
                => Cmd.none
                => NoOp

        AddProject projectJson ->
            let
                projects =
                    case Decode.decodeValue Project.decoder projectJson of
                        Ok project ->
                            addProject model.projects project

                        Err _ ->
                            model.projects
            in
                { model | projects = projects }
                    => Cmd.none
                    => NoOp
