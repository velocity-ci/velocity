module Page.Home
    exposing
        ( view
        , update
        , Model
        , Msg
        , ExternalMsg(..)
        , init
        , channelName
        , leaveChannels
        , initialEvents
        , subscriptions
        )

{-| The homepage. You can get here via either the / or /#/ routes.
-}

-- EXTERNAL

import Http
import Html exposing (..)
import Html.Attributes exposing (class, href, id, placeholder, attribute, classList, style, disabled)
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
import Component.KnownHostForm as KnownHostForm
import Component.Form as Form
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project, addProject)
import Data.KnownHost as KnownHost exposing (KnownHost, addKnownHost)
import Data.PaginatedList as PaginatedList exposing (Paginated(..))
import Data.GitUrl as GitUrl exposing (GitUrl)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Helpers exposing (formatDate, sortByDatetime)
import Page.Project.Route as ProjectRoute
import Views.Helpers exposing (onClickPage)
import Views.Page as Page
import Util exposing ((=>), onClickStopPropagation, viewIf)
import Request.Project
import Request.KnownHost
import Request.Errors
import Route
import Dom
import Ports


-- MODEL --


type alias Model =
    { projects : List Project
    , knownHosts : List KnownHost
    , newProjectForm : ProjectForm.Context
    , newKnownHostForm : KnownHostForm.Context
    , newProjectModalVisibility : Modal.Visibility
    }


init : Context -> Session msg -> Task (Request.Errors.Error PageLoadError) Model
init context session =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadProjects =
            Request.Project.list context maybeAuthToken

        loadKnownHosts =
            Request.KnownHost.list context maybeAuthToken

        errorPage =
            pageLoadError Page.Home "Homepage is currently unavailable."

        initialModel projects knownHosts =
            { projects = PaginatedList.results projects
            , knownHosts = PaginatedList.results knownHosts
            , newProjectForm = ProjectForm.init
            , newKnownHostForm = KnownHostForm.init
            , newProjectModalVisibility = Modal.hidden
            }
    in
        Task.map2 initialModel loadProjects loadKnownHosts
            |> Task.mapError (Request.Errors.withDefaultError errorPage)



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions { newProjectModalVisibility } =
    let
        modalSubs =
            Modal.subscriptions newProjectModalVisibility AnimateNewProjectModal

        gitUrlSub =
            Sub.map SetGitUrl gitUrlParsed
    in
        Sub.batch
            [ modalSubs, gitUrlSub ]


gitUrlParsed : Sub (Maybe GitUrl)
gitUrlParsed =
    Decode.decodeValue GitUrl.decoder
        >> Result.toMaybe
        |> Ports.onGitUrlParsed



-- CHANNELS --


channelName : String
channelName =
    "projects"


leaveChannels : Maybe Route.Route -> List String
leaveChannels route =
    case route of
        Just (Route.Home) ->
            []

        _ ->
            [ channelName ]


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
        div [ class "p-4 my-4" ]
            [ viewToolbar
            , viewIf hasProjects projectList
            , viewNewProjectModal model
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


knownHostFormConfig : Maybe GitUrl -> KnownHostForm.Config Msg
knownHostFormConfig maybeGitUrl =
    { setScannedKeyMsg = SetKnownHostFormScannedKey
    , submitMsg = SubmitKnownHostForm
    , gitUrl = maybeGitUrl
    }


viewNewProjectModal : Model -> Html Msg
viewNewProjectModal { newProjectForm, newKnownHostForm, newProjectModalVisibility, knownHosts } =
    Modal.config CloseNewProjectModal
        |> Modal.withAnimation AnimateNewProjectModal
        |> Modal.large
        |> Modal.hideOnBackdropClick (not newProjectForm.submitting)
        |> Modal.h3 [] [ text "Create project" ]
        |> Modal.body [] [ viewCombinedForm knownHosts newProjectForm newKnownHostForm ]
        |> Modal.footer [] [ viewCombinedFormSubmit knownHosts newProjectForm ]
        |> Modal.view newProjectModalVisibility


viewCombinedForm : List KnownHost -> ProjectForm.Context -> KnownHostForm.Context -> Html Msg
viewCombinedForm knownHosts projectForm knownHostForm =
    let
        projectFormView =
            ProjectForm.view projectFormConfig projectForm

        knownHostFormView =
            if ProjectForm.isUnknownHost knownHosts projectForm.form.gitUrl then
                KnownHostForm.view (knownHostFormConfig projectForm.form.gitUrl) knownHostForm
            else
                text ""
    in
        div []
            [ projectFormView
            , knownHostFormView
            ]


viewCombinedFormSubmit : List KnownHost -> ProjectForm.Context -> Html Msg
viewCombinedFormSubmit knownHosts projectForm =
    if not <| ProjectForm.isUnknownHost knownHosts projectForm.form.gitUrl then
        ProjectForm.viewSubmitButton projectFormConfig projectForm
    else
        Button.button
            [ Button.outlinePrimary
            , Button.attrs
                [ onClick SubmitBothForms
                  --            , disabled (hasErrors || submitting || untouched)
                ]
            ]
            [ text "Create" ]


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
    = NoOp_
    | NewUrl String
    | AddProject Encode.Value
    | AddKnownHost Encode.Value
    | CloseNewProjectModal
    | AnimateNewProjectModal Modal.Visibility
    | ShowNewProjectModal
    | SubmitProjectForm
    | SubmitBothForms
    | SetProjectFormName String
    | SetProjectFormRepository String
    | SetProjectFormPrivateKey String
    | SetKnownHostFormScannedKey String
    | SubmitKnownHostForm
    | ProjectCreated (Result Request.Errors.HttpError Project)
    | KnownHostCreated (Result Request.Errors.HttpError KnownHost)
    | KnownHostAndProjectCreated (Result Request.Errors.HttpError ( KnownHost, Project ))
    | SetGitUrl (Maybe GitUrl)


type ExternalMsg
    = NoOp
    | HandleRequestError Request.Errors.HandledError


formError : String -> List ( String, String )
formError errorMsg =
    [ "" => errorMsg ]


formErrors :
    Form.Context ProjectForm.Field projectForm
    -> Form.Context ProjectForm.Field projectForm
formErrors =
    ProjectForm.serverErrorToFormError
        |> Form.updateServerErrors (formError "Unable to process project.")


handleFormError :
    { a
        | newKnownHostForm : Form.Context KnownHostForm.Field knownHostForm
        , newProjectForm : Form.Context ProjectForm.Field projectForm
    }
    -> Request.Errors.Error Http.Error
    -> ( Form.Context ProjectForm.Field projectForm, Form.Context KnownHostForm.Field knownHostForm, ExternalMsg )
handleFormError model err =
    case err of
        Request.Errors.HandledError handledError ->
            ( model.newProjectForm, model.newKnownHostForm, HandleRequestError handledError )

        Request.Errors.UnhandledError (Http.BadStatus response) ->
            let
                projectErrors =
                    response.body
                        |> decodeString ProjectForm.errorsDecoder
                        |> Result.withDefault []

                projectForm =
                    model.newProjectForm
                        |> Form.updateServerErrors projectErrors ProjectForm.serverErrorToFormError

                knownHostErrors =
                    response.body
                        |> decodeString KnownHostForm.errorsDecoder
                        |> Result.withDefault []

                knownHostForm =
                    model.newKnownHostForm
                        |> Form.updateServerErrors knownHostErrors KnownHostForm.serverErrorToFormError
            in
                ( projectForm, knownHostForm, NoOp )

        _ ->
            ( formErrors model.newProjectForm, model.newKnownHostForm, NoOp )


update : Context -> Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update context session msg model =
    case msg of
        NoOp_ ->
            model
                => Cmd.none
                => NoOp

        NewUrl url ->
            model
                => newUrl url
                => NoOp

        SetGitUrl maybeGitUrl ->
            { model | newProjectForm = ProjectForm.updateGitUrl maybeGitUrl model.newProjectForm }
                => Cmd.none
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
                => Ports.parseGitUrl repository
                => NoOp

        SetProjectFormPrivateKey privateKey ->
            { model | newProjectForm = ProjectForm.update model.newProjectForm ProjectForm.PrivateKey privateKey }
                => Cmd.none
                => NoOp

        SubmitBothForms ->
            let
                submitBothForms authToken =
                    authToken
                        |> Request.KnownHost.create context (KnownHostForm.submitValues model.newKnownHostForm)
                        |> Task.andThen
                            (\knownHostResult ->
                                Request.Project.create context (ProjectForm.submitValues model.newProjectForm) authToken
                                    |> Task.andThen (\projectResult -> Task.succeed ( knownHostResult, projectResult ))
                            )
            in
                case Maybe.map .token session.user of
                    Just authToken ->
                        model
                            => Task.attempt KnownHostAndProjectCreated (submitBothForms authToken)
                            => NoOp

                    Nothing ->
                        model
                            => Cmd.none
                            => NoOp

        KnownHostAndProjectCreated (Ok ( knownHost, project )) ->
            { model
                | newProjectForm = ProjectForm.init
                , newProjectModalVisibility = Modal.hidden
                , projects = addProject model.projects project
                , knownHosts = addKnownHost model.knownHosts knownHost
            }
                => Cmd.none
                => NoOp

        KnownHostAndProjectCreated (Err err) ->
            let
                ( updatedProjectForm, updatedKnownHostForm, externalMsg ) =
                    handleFormError model err
            in
                { model
                    | newProjectForm = Form.submitting False updatedProjectForm
                    , newKnownHostForm = Form.submitting False updatedKnownHostForm
                }
                    => Cmd.none
                    => externalMsg

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

        SubmitKnownHostForm ->
            let
                cmdFromAuth authToken =
                    authToken
                        |> Request.KnownHost.create context (KnownHostForm.submitValues model.newKnownHostForm)
                        |> Task.attempt KnownHostCreated

                cmd =
                    session
                        |> Session.attempt "create known host" cmdFromAuth
                        |> Tuple.second
            in
                { model | newKnownHostForm = Form.submit model.newKnownHostForm }
                    => cmd
                    => NoOp

        SetKnownHostFormScannedKey key ->
            { model | newKnownHostForm = KnownHostForm.update model.newKnownHostForm KnownHostForm.ScannedKey key }
                => Cmd.none
                => NoOp

        ProjectCreated (Err err) ->
            let
                ( updatedProjectForm, _, externalMsg ) =
                    handleFormError model err
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

        KnownHostCreated (Err err) ->
            let
                ( _, updatedKnownHostForm, externalMsg ) =
                    handleFormError model err
            in
                { model | newKnownHostForm = Form.submitting False updatedKnownHostForm }
                    => Cmd.none
                    => externalMsg

        KnownHostCreated (Ok knownHost) ->
            let
                updatedModel =
                    { model
                        | newKnownHostForm = KnownHostForm.init
                        , knownHosts = addKnownHost model.knownHosts knownHost
                    }
            in
                updatedModel
                    => Cmd.none
                    => NoOp

        ShowNewProjectModal ->
            { model | newProjectModalVisibility = Modal.shown }
                => Task.attempt (always NoOp_) (Dom.focus "name")
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

        AddKnownHost knownHostJson ->
            let
                knownHosts =
                    case Decode.decodeValue KnownHost.decoder knownHostJson of
                        Ok knownHost ->
                            addKnownHost model.knownHosts knownHost

                        Err _ ->
                            model.knownHosts
            in
                { model | knownHosts = knownHosts }
                    => Cmd.none
                    => NoOp
