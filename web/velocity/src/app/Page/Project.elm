module Page.Project exposing (..)

import Context exposing (Context)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.Project
import Request.Errors
import Task exposing (Task)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Data.Branch as Branch exposing (Branch)
import Data.Event as Event exposing (Event)
import Data.Task as ProjectTask
import Html exposing (..)
import Html.Attributes exposing (..)
import Views.Page as Page
import Util exposing ((=>), viewIf)
import Http
import Route exposing (Route)
import Navigation exposing (newUrl)
import Data.PaginatedList as PaginatedList exposing (PaginatedList)
import Json.Encode as Encode
import Json.Decode as Decode
import Dict exposing (Dict)
import Data.Build as Build exposing (Build, addBuild)
import Page.Project.Route as ProjectRoute
import Page.Project.Commits as Commits
import Page.Project.Settings as Settings
import Page.Project.Commit as Commit
import Page.Project.Builds as Builds
import Page.Helpers exposing (sortByDatetime)
import Bootstrap.Popover as Popover
import Component.ProjectSidebar as Sidebar exposing (ActiveSubPage(..))
import Toasty
import Views.Build
import Views.Toast as ToastTheme
import Views.Page as Page exposing (ActivePage)
import Views.Helpers exposing (onClickPage)


-- SUB PAGES --


type SubPage
    = Blank
    | Commits Commits.Model
    | Commit Commit.Model
    | Settings Settings.Model
    | Builds Builds.Model
    | Errored PageLoadError


type SubPageState
    = Loaded SubPage
    | TransitioningFrom SubPage



-- MODEL --


type alias Model =
    { project : Project
    , branches : List Branch
    , subPageState : SubPageState
    , sidebar : Sidebar.State
    , toasties : Toasty.Stack (Event Build)
    }


initialSubPage : SubPage
initialSubPage =
    Blank


init : Context -> Session msg -> Project.Slug -> Maybe ProjectRoute.Route -> Task (Request.Errors.Error PageLoadError) ( Model, Cmd Msg )
init context session slug maybeRoute =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadProject =
            Request.Project.get context slug maybeAuthToken

        loadBranches =
            Request.Project.branches context slug maybeAuthToken

        initialModel project paginatedBranches =
            { project = project
            , branches = PaginatedList.results paginatedBranches
            , subPageState = Loaded initialSubPage
            , sidebar =
                { commitIconPopover = Popover.initialState
                , buildsIconPopover = Popover.initialState
                , settingsIconPopover = Popover.initialState
                , projectBadgePopover = Popover.initialState
                }
            , toasties = Toasty.initialState
            }

        handleLoadError e =
            case e of
                Http.BadPayload debugError _ ->
                    pageLoadError Page.Project debugError

                _ ->
                    pageLoadError Page.Project "Project unavailable."
    in
        Task.map2 initialModel loadProject loadBranches
            |> Task.map
                (\successModel ->
                    case maybeRoute of
                        Just route ->
                            update context session (SetRoute maybeRoute) successModel

                        Nothing ->
                            ( successModel, Cmd.none )
                )
            |> Task.mapError (Request.Errors.mapUnhandledError handleLoadError)



-- CHANNELS --


channelName : Project.Slug -> String
channelName projectSlug =
    "project:" ++ (Project.slugToString projectSlug)


initialEvents : Project.Slug -> ProjectRoute.Route -> Dict String (List ( String, Encode.Value -> Msg ))
initialEvents slug route =
    let
        mapEvents fromMsg events =
            events
                |> Dict.map (\_ v -> List.map (Tuple.mapSecond (\msg -> msg >> fromMsg)) v)

        subPageEvents =
            case route of
                ProjectRoute.Commits _ _ ->
                    mapEvents CommitsMsg (Commits.events slug)

                ProjectRoute.Commit _ subRoute ->
                    mapEvents CommitMsg (Commit.initialEvents (channelName slug) subRoute)

                _ ->
                    Dict.empty

        pageEvents =
            [ ( "project:update", UpdateProject )
            , ( "project:delete", ProjectDeleted )
            , ( "branch:new", RefreshBranches )
            , ( "branch:update", RefreshBranches )
            , ( "branch:delete", RefreshBranches )
            , ( "build:new", AddBuildEvent )
            , ( "build:delete", DeleteBuildEvent )
            , ( "build:update", UpdateBuildEvent )
            ]

        merge pageDict =
            let
                existsInPage =
                    Dict.insert

                existsInSubPage =
                    Dict.insert

                existsInBoth key a b dict =
                    Dict.insert key (List.append a b) dict
            in
                Dict.merge existsInPage existsInBoth existsInSubPage pageDict subPageEvents Dict.empty
    in
        Dict.singleton (channelName slug) pageEvents
            |> merge


loadedEvents : Msg -> Model -> Dict String (List ( String, Encode.Value -> Msg ))
loadedEvents msg model =
    case ( msg, getSubPage model.subPageState ) of
        ( CommitMsg subMsg, Commit subModel ) ->
            Commit.loadedEvents subMsg subModel
                |> Dict.map (\_ v -> List.map (Tuple.mapSecond (\msg -> msg >> CommitMsg)) v)

        _ ->
            Dict.empty


leaveChannels : Model -> Maybe Route.Route -> List String
leaveChannels model route =
    let
        projectChannel =
            channelName model.project.slug
    in
        case route of
            Just (Route.Project slug subRoute) ->
                let
                    subPageChannels =
                        leaveSubPageChannels (getSubPage model.subPageState) subRoute
                in
                    if slug == model.project.slug then
                        -- Route is for this project. Maybe different sub route.
                        subPageChannels
                    else
                        -- Project route, but for a different project
                        projectChannel :: subPageChannels

            -- Not a project route
            _ ->
                [ projectChannel ]


leaveSubPageChannels : SubPage -> ProjectRoute.Route -> List String
leaveSubPageChannels subPage subRoute =
    case subPage of
        Commit subModel ->
            Commit.leaveChannels subModel subRoute

        _ ->
            []



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    case (getSubPage model.subPageState) of
        Commits subModel ->
            Commits.subscriptions model.branches subModel
                |> Sub.map CommitsMsg

        Commit subModel ->
            Commit.subscriptions model.project subModel
                |> Sub.map CommitMsg

        _ ->
            Sub.none



-- VIEW --


view : Session msg -> Model -> Html Msg
view session model =
    div [ class "p-2" ]
        [ viewSubPage session model
        , toastView model
        ]


toastView : Model -> Html Msg
toastView { toasties } =
    Toasty.view ToastTheme.config Views.Build.toast ToastyMsg toasties


sidebarConfig : Sidebar.Config Msg
sidebarConfig =
    { newUrlMsg = NewUrl
    , commitPopMsg = CommitsIconPopMsg
    , buildsPopMsg = BuildsIconPopMsg
    , settingsPopMsg = SettingsIconPopMsg
    , projectBadgePopMsg = ProjectBadgePopMsg
    }


viewSidebar : Session msg -> Model -> Html Msg
viewSidebar session model =
    let
        page =
            getSubPage model.subPageState

        sidebar =
            Sidebar.view model.sidebar sidebarConfig model.project
    in
        case page of
            Commits _ ->
                sidebar CommitsPage

            Commit _ ->
                sidebar CommitsPage

            Settings _ ->
                sidebar SettingsPage

            Builds _ ->
                sidebar BuildsPage

            _ ->
                sidebar OtherPage


viewSubPage : Session msg -> Model -> Html Msg
viewSubPage session model =
    let
        page =
            getSubPage model.subPageState

        project =
            model.project

        branches =
            model.branches

        breadcrumb =
            viewBreadcrumb project

        pageFrame =
            frame project model
    in
        case page of
            Blank ->
                Html.text ""
                    |> pageFrame (breadcrumb (text "") [])

            Errored subModel ->
                Html.text "Errored"
                    |> pageFrame (breadcrumb (text "") [])

            Builds subModel ->
                Builds.view project subModel
                    |> Html.map BuildsMsg
                    |> pageFrame (breadcrumb (text "") [])

            Commits subModel ->
                let
                    crumb =
                        Commits.breadcrumb project
                            |> breadcrumb crumbExtraItems

                    crumbExtraItems =
                        Commits.viewBreadcrumbExtraItems project subModel
                            |> Html.map CommitsMsg
                in
                    Commits.view project branches subModel
                        |> Html.map CommitsMsg
                        |> pageFrame crumb

            Commit subModel ->
                let
                    crumb =
                        Commit.breadcrumb project subModel.commit subModel.subPageState
                            |> breadcrumb (text "")
                in
                    Commit.view project session subModel
                        |> Html.map CommitMsg
                        |> pageFrame crumb

            Settings subModel ->
                let
                    crumb =
                        Settings.breadcrumb project
                            |> breadcrumb (text "")
                in
                    Settings.view project subModel
                        |> Html.map SettingsMsg
                        |> pageFrame crumb


frame : Project -> Model -> Html Msg -> Html Msg -> Html Msg
frame project model breadcrumb content =
    div []
        [ topNav breadcrumb (subNavbar model)
        , div [ class "px-3 py-3" ]
            [ content
            ]
        ]


topNav : Html Msg -> Html Msg -> Html Msg
topNav breadcrumb subNav =
    nav [ class "navbar navbar-expand-lg bg-white navbar-light" ]
        [ subNav
        , breadcrumb
        ]


subNavbar : Model -> Html Msg
subNavbar model =
    case getSubPage model.subPageState of
        Commit subModel ->
            Commit.viewNavbar subModel
                |> Html.map CommitMsg

        _ ->
            text ""



-- BREADCRUMB


viewBreadcrumb : Project -> Html Msg -> List ( Route, String ) -> Html Msg
viewBreadcrumb project additionalElements items =
    let
        fixedItems =
            [ ( Route.Project project.slug ProjectRoute.Overview, project.name )
            ]

        allItems =
            fixedItems ++ items

        allItemLength =
            List.length allItems

        breadcrumbItem i item =
            item |> viewBreadcrumbItem (i == (allItemLength - 1))

        itemElements =
            allItems
                |> List.indexedMap breadcrumbItem
    in
        div []
            [ ol [ class "px-0 breadcrumb bg-white mb-2 pb-0" ] itemElements
            , additionalElements
            ]


viewBreadcrumbItem : Bool -> ( Route, String ) -> Html Msg
viewBreadcrumbItem active ( route, name ) =
    Util.viewIf (not active) <|
        li
            [ Route.href route
            , onClickPage NewUrl route
            , class "breadcrumb-item"
            , classList [ ( "active", active ) ]
            ]
            [ a [ Route.href route, class "text-secondary" ] [ text name ] ]



-- UPDATE --


type Msg
    = NewUrl String
    | SetRoute (Maybe ProjectRoute.Route)
    | CommitsMsg Commits.Msg
    | CommitsLoaded (Result PageLoadError Commits.Model)
    | BuildsMsg Builds.Msg
    | BuildsLoaded (Result PageLoadError Builds.Model)
    | CommitMsg Commit.Msg
    | CommitLoaded (Result PageLoadError ( Commit.Model, Cmd Commit.Msg ))
    | SettingsMsg Settings.Msg
    | UpdateProject Encode.Value
    | AddBranch Encode.Value
    | ProjectDeleted Encode.Value
    | RefreshBranches Encode.Value
    | RefreshBranchesComplete (Result Request.Errors.HttpError (PaginatedList Branch))
    | AddBuildEvent Encode.Value
    | UpdateBuildEvent Encode.Value
    | DeleteBuildEvent Encode.Value
    | CommitsIconPopMsg Popover.State
    | BuildsIconPopMsg Popover.State
    | SettingsIconPopMsg Popover.State
    | ProjectBadgePopMsg Popover.State
    | ToastyMsg (Toasty.Msg (Event Build))
    | NoOp


getSubPage : SubPageState -> SubPage
getSubPage subPageState =
    case subPageState of
        Loaded subPage ->
            subPage

        TransitioningFrom subPage ->
            subPage


setRoute : Context -> Session msg -> Maybe ProjectRoute.Route -> Model -> ( Model, Cmd Msg )
setRoute context session maybeRoute model =
    let
        transition toMsg task =
            { model | subPageState = TransitioningFrom (getSubPage model.subPageState) }
                => Task.attempt toMsg task

        errored =
            pageErrored model
    in
        case maybeRoute of
            Nothing ->
                { model | subPageState = Loaded Blank } => Cmd.none

            Just (ProjectRoute.Overview) ->
                model => Route.modifyUrl (Route.Project model.project.slug <| ProjectRoute.Commits Nothing Nothing)

            Just (ProjectRoute.Commits maybeBranch maybePage) ->
                case session.user of
                    Just user ->
                        Commits.init context session model.branches model.project.slug maybeBranch maybePage
                            |> transition CommitsLoaded

                    Nothing ->
                        errored Page.Project "Uhoh"

            Just (ProjectRoute.Commit hash maybeRoute) ->
                let
                    loadFreshPage =
                        Just maybeRoute
                            |> Commit.init context session model.project hash
                            |> transition CommitLoaded

                    transitionSubPage subModel =
                        let
                            ( newModel, newMsg ) =
                                subModel
                                    |> Commit.update context model.project session (Commit.SetRoute (Just maybeRoute))
                        in
                            { model | subPageState = Loaded (Commit newModel) }
                                => Cmd.map CommitMsg newMsg
                in
                    case ( session.user, model.subPageState ) of
                        ( Just _, Loaded page ) ->
                            case page of
                                -- If we're on the product page for the same product as the new route just load sub-page
                                -- Otherwise load the project page fresh
                                Commit subModel ->
                                    if hash == subModel.commit.hash then
                                        transitionSubPage subModel
                                    else
                                        loadFreshPage

                                _ ->
                                    loadFreshPage

                        ( Just _, TransitioningFrom _ ) ->
                            loadFreshPage

                        ( Nothing, _ ) ->
                            errored Page.Project "Error loading commit"

            Just (ProjectRoute.Builds maybePage) ->
                case session.user of
                    Just user ->
                        maybePage
                            |> Builds.init context session model.project.slug
                            |> transition BuildsLoaded

                    Nothing ->
                        errored Page.Project "Uhoh"

            Just (ProjectRoute.Settings) ->
                case session.user of
                    Just user ->
                        { model | subPageState = Loaded (Settings (Settings.initialModel)) } => Cmd.none

                    Nothing ->
                        errored Page.Project "Uhoh"


pageErrored : Model -> ActivePage -> String -> ( Model, Cmd msg )
pageErrored model activePage errorMessage =
    let
        error =
            Errored.pageLoadError activePage errorMessage
    in
        { model | subPageState = Loaded (Errored error) } => Cmd.none


update : Context -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update context session msg model =
    case updateSidebar model.sidebar msg of
        Just sidebar ->
            { model | sidebar = sidebar }
                => Cmd.none

        Nothing ->
            updateSubPage context session (getSubPage model.subPageState) msg model


updateSidebar : Sidebar.State -> Msg -> Maybe Sidebar.State
updateSidebar sidebar msg =
    case msg of
        CommitsIconPopMsg state ->
            Just { sidebar | commitIconPopover = state }

        SettingsIconPopMsg state ->
            Just { sidebar | settingsIconPopover = state }

        BuildsIconPopMsg state ->
            Just { sidebar | buildsIconPopover = state }

        ProjectBadgePopMsg state ->
            Just { sidebar | projectBadgePopover = state }

        _ ->
            Nothing


updateSubPage : Context -> Session msg -> SubPage -> Msg -> Model -> ( Model, Cmd Msg )
updateSubPage context session subPage msg model =
    let
        toPage toModel toMsg subUpdate subMsg subModel =
            let
                ( newModel, newCmd ) =
                    subUpdate subMsg subModel
            in
                ( { model | subPageState = Loaded (toModel newModel) }, Cmd.map toMsg newCmd )

        sidebar =
            model.sidebar

        errored =
            pageErrored model
    in
        case ( msg, subPage ) of
            ( NoOp, _ ) ->
                model => Cmd.none

            ( NewUrl url, _ ) ->
                model
                    => newUrl url

            ( SetRoute route, _ ) ->
                setRoute context session route model

            ( SettingsMsg subMsg, Settings subModel ) ->
                toPage Settings SettingsMsg (Settings.update context model.project session) subMsg subModel

            ( CommitsLoaded (Ok subModel), _ ) ->
                { model | subPageState = Loaded (Commits subModel) }
                    => Cmd.none

            ( CommitsLoaded (Err error), _ ) ->
                { model | subPageState = Loaded (Errored error) }
                    => Cmd.none

            ( CommitsMsg subMsg, Commits subModel ) ->
                toPage Commits CommitsMsg (Commits.update context model.project session) subMsg subModel

            ( CommitLoaded (Ok ( subModel, subMsg )), _ ) ->
                { model | subPageState = Loaded (Commit subModel) }
                    => Cmd.map CommitMsg subMsg

            ( CommitLoaded (Err error), _ ) ->
                { model | subPageState = Loaded (Errored (Debug.log "ERROR " error)) }
                    => Cmd.none

            ( CommitMsg subMsg, Commit subModel ) ->
                let
                    ( newSubModel, newCmd ) =
                        Commit.update context model.project session subMsg subModel
                in
                    { model | subPageState = Loaded (Commit newSubModel) }
                        ! [ Cmd.map CommitMsg newCmd ]

            ( BuildsMsg subMsg, Builds subModel ) ->
                toPage Builds BuildsMsg (Builds.update context model.project session) subMsg subModel

            ( BuildsLoaded (Ok subModel), _ ) ->
                { model | subPageState = Loaded (Builds subModel) }
                    => Cmd.none

            ( BuildsLoaded (Err error), _ ) ->
                { model | subPageState = Loaded (Errored error) }
                    => Cmd.none

            ( UpdateProject updateJson, _ ) ->
                let
                    newProject =
                        updateJson
                            |> Decode.decodeValue Project.decoder
                            |> Result.toMaybe
                            |> Maybe.withDefault model.project
                in
                    { model | project = newProject }
                        => Cmd.none

            ( AddBranch branchJson, _ ) ->
                let
                    branches =
                        Decode.decodeValue Branch.decoder branchJson
                            |> Result.toMaybe
                            |> Maybe.map (\b -> b :: model.branches)
                            |> Maybe.withDefault model.branches
                in
                    { model | branches = branches }
                        => Cmd.none

            ( RefreshBranches _, _ ) ->
                let
                    cmd =
                        Request.Project.branches context model.project.slug (Maybe.map .token session.user)
                            |> Task.attempt RefreshBranchesComplete
                in
                    model => cmd

            ( RefreshBranchesComplete (Ok paginatedBranches), _ ) ->
                { model | branches = PaginatedList.results paginatedBranches }
                    => Cmd.none

            ( ProjectDeleted _, _ ) ->
                model
                    => Route.modifyUrl Route.Projects

            ( AddBuildEvent buildJson, page ) ->
                let
                    maybeBuild =
                        buildJson
                            |> Decode.decodeValue Build.decoder
                            |> Result.toMaybe

                    ( subPageState, subCmd ) =
                        case ( page, maybeBuild ) of
                            ( Commit subModel, Just build ) ->
                                subModel
                                    |> Commit.update context model.project session (Commit.AddBuild build)
                                    |> Tuple.mapFirst (Commit >> Loaded)
                                    |> Tuple.mapSecond (Cmd.map CommitMsg)

                            ( _, _ ) ->
                                model.subPageState => Cmd.none

                    modelCmd =
                        { model | subPageState = subPageState }
                            ! [ subCmd ]
                in
                    case maybeBuild of
                        Just build ->
                            Toasty.addToast ToastTheme.config ToastyMsg (Event.Created build) modelCmd

                        Nothing ->
                            modelCmd

            ( UpdateBuildEvent buildJson, page ) ->
                let
                    maybeBuild =
                        buildJson
                            |> Decode.decodeValue Build.decoder
                            |> Result.toMaybe

                    ( subPageState, subCmd ) =
                        case ( page, maybeBuild ) of
                            ( Commit subModel, Just build ) ->
                                subModel
                                    |> Commit.update context model.project session (Commit.UpdateBuild build)
                                    |> Tuple.mapFirst (Commit >> Loaded)
                                    |> Tuple.mapSecond (Cmd.map CommitMsg)

                            ( _, _ ) ->
                                model.subPageState => Cmd.none

                    modelCmd =
                        { model | subPageState = subPageState }
                            ! [ subCmd ]

                    modelCmdWithToast =
                        case maybeBuild of
                            Just build ->
                                Toasty.addToast ToastTheme.config ToastyMsg (Event.Completed build) modelCmd

                            Nothing ->
                                modelCmd
                in
                    case Maybe.map .status maybeBuild of
                        Just (Build.Success) ->
                            modelCmdWithToast

                        Just (Build.Failed) ->
                            modelCmdWithToast

                        _ ->
                            modelCmd

            ( DeleteBuildEvent buildJson, _ ) ->
                model => Cmd.none

            ( ToastyMsg subMsg, _ ) ->
                Toasty.update ToastTheme.config ToastyMsg subMsg model

            ( _, _ ) ->
                -- Disregard incoming messages that arrived for the wrong sub page
                (Debug.log "Fell through (project page)" model)
                    => Cmd.none



-- HELPERS --


hasExtraWideSidebar : Model -> Session msg -> Bool
hasExtraWideSidebar { subPageState } session =
    case getSubPage subPageState of
        Commit subModel ->
            Commit.hasExtraWideSidebar subModel

        _ ->
            False
