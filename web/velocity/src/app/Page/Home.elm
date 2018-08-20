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
    , forms : Forms
    , newProjectModalVisibility : Modal.Visibility
    }


type Forms
    = ProjectFormOnly ProjectForm.Context
    | ProjectAndKnownHostForm ProjectForm.Context KnownHostForm.Context


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
            , forms = ProjectFormOnly ProjectForm.init
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
    , submitMsg = NoOp_
    }


knownHostFormConfig : Maybe GitUrl -> KnownHostForm.Config Msg
knownHostFormConfig maybeGitUrl =
    { setScannedKeyMsg = SetKnownHostFormScannedKey
    , submitMsg = NoOp_
    , gitUrl = maybeGitUrl
    }


viewNewProjectModal : Model -> Html Msg
viewNewProjectModal { forms, newProjectModalVisibility, knownHosts } =
    Modal.config CloseNewProjectModal
        |> Modal.withAnimation AnimateNewProjectModal
        |> Modal.large
        |> Modal.hideOnBackdropClick (not (forms |> projectForm |> .submitting))
        |> Modal.h3 [] [ text "Create project" ]
        |> Modal.body [] [ viewCombinedForm knownHosts forms ]
        |> Modal.footer [] [ viewCombinedFormSubmit knownHosts forms ]
        |> Modal.view newProjectModalVisibility


projectForm : Forms -> ProjectForm.Context
projectForm forms =
    case forms of
        ProjectFormOnly projectForm ->
            projectForm

        ProjectAndKnownHostForm projectForm _ ->
            projectForm


viewCombinedForm : List KnownHost -> Forms -> Html Msg
viewCombinedForm knownHosts forms =
    case forms of
        ProjectFormOnly projectForm ->
            div []
                [ ProjectForm.view projectFormConfig projectForm ]

        ProjectAndKnownHostForm projectForm knownHostForm ->
            div []
                [ ProjectForm.view projectFormConfig projectForm
                , KnownHostForm.view (knownHostFormConfig projectForm.form.gitUrl) knownHostForm
                ]


viewCombinedFormSubmit : List KnownHost -> Forms -> Html Msg
viewCombinedFormSubmit knownHosts forms =
    Button.button
        [ Button.outlinePrimary
        , Button.attrs
            [ onClick SubmitBothForms
            , disabled (hasErrors forms || submitting forms || untouched forms)
            ]
        ]
        [ text "Create" ]


hasErrors : Forms -> Bool
hasErrors forms =
    case forms of
        ProjectFormOnly projectForm ->
            not (List.isEmpty projectForm.errors)

        ProjectAndKnownHostForm projectForm knownHostForm ->
            not (List.isEmpty projectForm.errors) || not (List.isEmpty knownHostForm.errors)


submitting : Forms -> Bool
submitting forms =
    case forms of
        ProjectFormOnly projectForm ->
            projectForm.submitting

        ProjectAndKnownHostForm projectForm knownHostForm ->
            projectForm.submitting || knownHostForm.submitting


untouched : Forms -> Bool
untouched forms =
    case forms of
        ProjectFormOnly projectForm ->
            ProjectForm.isUntouched projectForm

        ProjectAndKnownHostForm projectForm knownHostForm ->
            ProjectForm.isUntouched projectForm && KnownHostForm.isUntouched knownHostForm


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
    | SubmitBothForms
    | SetProjectFormName String
    | SetProjectFormRepository String
    | SetProjectFormPrivateKey String
    | SetKnownHostFormScannedKey String
    | ProjectCreated (Result Request.Errors.HttpError Project)
    | KnownHostAndProjectCreated (Result Request.Errors.HttpError ( KnownHost, Project ))
    | SetGitUrl (Maybe GitUrl)


type ExternalMsg
    = NoOp
    | HandleRequestError Request.Errors.HandledError


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
            let
                updatedProjectForm =
                    model.forms
                        |> projectForm
                        |> ProjectForm.updateGitUrl maybeGitUrl

                isUnknownHost =
                    ProjectForm.isUnknownHost model.knownHosts updatedProjectForm.form.gitUrl

                isSshAddress =
                    ProjectForm.isSshAddress updatedProjectForm.form.gitUrl

                needsKnownHostForm =
                    isUnknownHost && isSshAddress
            in
                { model
                    | forms =
                        case ( needsKnownHostForm, model.forms ) of
                            ( True, ProjectFormOnly _ ) ->
                                ProjectAndKnownHostForm updatedProjectForm KnownHostForm.init

                            ( True, ProjectAndKnownHostForm _ knownHostForm ) ->
                                ProjectAndKnownHostForm updatedProjectForm knownHostForm

                            ( _, _ ) ->
                                ProjectFormOnly updatedProjectForm
                }
                    => Cmd.none
                    => NoOp

        CloseNewProjectModal ->
            { model
                | newProjectModalVisibility = Modal.hidden
                , forms = ProjectFormOnly ProjectForm.init
            }
                => Cmd.none
                => NoOp

        AnimateNewProjectModal visibility ->
            { model | newProjectModalVisibility = visibility }
                => Cmd.none
                => NoOp

        SetProjectFormName name ->
            let
                updatedProjectForm =
                    ProjectForm.update (projectForm model.forms) ProjectForm.Name name
            in
                { model
                    | forms =
                        case model.forms of
                            ProjectFormOnly _ ->
                                ProjectFormOnly updatedProjectForm

                            ProjectAndKnownHostForm _ knownHostForm ->
                                ProjectAndKnownHostForm updatedProjectForm knownHostForm
                }
                    => Cmd.none
                    => NoOp

        SetProjectFormRepository repository ->
            let
                updatedProjectForm =
                    ProjectForm.update (projectForm model.forms) ProjectForm.Repository repository
            in
                { model
                    | forms =
                        case model.forms of
                            ProjectFormOnly _ ->
                                ProjectFormOnly updatedProjectForm

                            ProjectAndKnownHostForm _ knownHostForm ->
                                ProjectAndKnownHostForm updatedProjectForm knownHostForm
                }
                    => Ports.parseGitUrl repository
                    => NoOp

        SetProjectFormPrivateKey privateKey ->
            let
                updatedProjectForm =
                    ProjectForm.update (projectForm model.forms) ProjectForm.PrivateKey privateKey
            in
                { model
                    | forms =
                        case model.forms of
                            ProjectFormOnly _ ->
                                ProjectFormOnly updatedProjectForm

                            ProjectAndKnownHostForm _ knownHostForm ->
                                ProjectAndKnownHostForm updatedProjectForm knownHostForm
                }
                    => Cmd.none
                    => NoOp

        SubmitBothForms ->
            let
                submit authToken =
                    case model.forms of
                        ProjectFormOnly projectForm ->
                            Request.Project.create context (ProjectForm.submitValues projectForm) authToken
                                |> Task.attempt ProjectCreated

                        ProjectAndKnownHostForm projectForm knownHostForm ->
                            authToken
                                |> Request.KnownHost.create context (KnownHostForm.submitValues knownHostForm)
                                |> Task.andThen
                                    (\knownHostResult ->
                                        Request.Project.create context (ProjectForm.submitValues projectForm) authToken
                                            |> Task.andThen (\projectResult -> Task.succeed ( knownHostResult, projectResult ))
                                    )
                                |> Task.attempt KnownHostAndProjectCreated
            in
                case Maybe.map .token session.user of
                    Just authToken ->
                        model
                            => submit authToken
                            => NoOp

                    Nothing ->
                        model
                            => Cmd.none
                            => NoOp

        KnownHostAndProjectCreated (Ok ( knownHost, project )) ->
            { model
                | forms = ProjectFormOnly ProjectForm.init
                , newProjectModalVisibility = Modal.hidden
                , projects = addProject model.projects project
                , knownHosts = addKnownHost model.knownHosts knownHost
            }
                => Route.modifyUrl (Route.Project project.slug ProjectRoute.default)
                => NoOp

        KnownHostAndProjectCreated (Err err) ->
            let
                ( forms, externalMsg ) =
                    handleFormError model err
            in
                { model | forms = resetSubmittingForms forms }
                    => Cmd.none
                    => externalMsg

        SetKnownHostFormScannedKey key ->
            let
                forms =
                    case model.forms of
                        ProjectFormOnly projectForm ->
                            ProjectFormOnly projectForm

                        ProjectAndKnownHostForm projectForm knownHostForm ->
                            KnownHostForm.update knownHostForm KnownHostForm.ScannedKey key
                                |> ProjectAndKnownHostForm projectForm
            in
                { model | forms = forms }
                    => Cmd.none
                    => NoOp

        ProjectCreated (Err err) ->
            let
                ( forms, externalMsg ) =
                    handleFormError model err
            in
                { model | forms = resetSubmittingForms forms }
                    => Cmd.none
                    => externalMsg

        ProjectCreated (Ok project) ->
            { model
                | forms = ProjectFormOnly ProjectForm.init
                , newProjectModalVisibility = Modal.hidden
                , projects = addProject model.projects project
            }
                => Route.modifyUrl (Route.Project project.slug ProjectRoute.default)
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


formError : String -> List ( String, String )
formError errorMsg =
    [ "" => errorMsg ]


formErrors :
    Form.Context ProjectForm.Field projectForm
    -> Form.Context ProjectForm.Field projectForm
formErrors =
    ProjectForm.serverErrorToFormError
        |> Form.updateServerErrors (formError "Unable to process project.")


handleFormError : Model -> Request.Errors.Error Http.Error -> ( Forms, ExternalMsg )
handleFormError model err =
    case err of
        Request.Errors.HandledError handledError ->
            ( model.forms, HandleRequestError handledError )

        Request.Errors.UnhandledError (Http.BadStatus response) ->
            let
                projectErrors =
                    response.body
                        |> decodeString ProjectForm.errorsDecoder
                        |> Result.withDefault []

                projectFormWithErrors =
                    model.forms
                        |> projectForm
                        |> Form.updateServerErrors projectErrors ProjectForm.serverErrorToFormError
            in
                case model.forms of
                    ProjectFormOnly _ ->
                        ProjectFormOnly projectFormWithErrors
                            => NoOp

                    ProjectAndKnownHostForm _ knownHostForm ->
                        let
                            knownHostErrors =
                                response.body
                                    |> decodeString KnownHostForm.errorsDecoder
                                    |> Result.withDefault []

                            knownHostFormWithErrors =
                                knownHostForm
                                    |> Form.updateServerErrors knownHostErrors KnownHostForm.serverErrorToFormError
                        in
                            ProjectAndKnownHostForm projectFormWithErrors knownHostFormWithErrors
                                => NoOp

        _ ->
            case model.forms of
                ProjectFormOnly projectForm ->
                    ProjectFormOnly (formErrors projectForm)
                        => NoOp

                ProjectAndKnownHostForm projectForm knownHostForm ->
                    ProjectAndKnownHostForm (formErrors projectForm) knownHostForm
                        => NoOp


resetSubmittingForms : Forms -> Forms
resetSubmittingForms forms =
    case forms of
        ProjectFormOnly projectForm ->
            ProjectFormOnly (Form.submitting False projectForm)

        ProjectAndKnownHostForm projectForm knownHostForm ->
            ProjectAndKnownHostForm (Form.submitting False projectForm) (Form.submitting False knownHostForm)
