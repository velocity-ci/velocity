module Page.Project.Commit.Task exposing (..)

-- EXTERNAL --

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, on, onSubmit)
import Task exposing (Task)
import Dict exposing (Dict)
import Json.Encode as Encode
import Array exposing (Array)
import Bootstrap.Modal as Modal
import Navigation


-- INTERNAL --

import Context exposing (Context)
import Component.BuildOutput as BuildOutput
import Component.BuildForm as BuildForm
import Data.AuthToken as AuthToken exposing (AuthToken)
import Data.Commit as Commit exposing (Commit)
import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Data.Build as Build exposing (Build)
import Data.BuildStream as BuildStream exposing (Id, BuildStream, BuildStreamOutput)
import Data.Task as ProjectTask exposing (Step(..), Parameter(..))
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Util exposing ((=>))
import Views.Page as Page
import Views.Task exposing (viewStepList)
import Views.Helpers exposing (onClickPage)
import Views.Build exposing (viewBuildStatusIcon, viewBuildStepStatusIcon, viewBuildTextClass)
import Request.Commit
import Request.Errors
import Route


-- MODEL --


type alias Model =
    { task : ProjectTask.Task
    , toggledStep : Maybe Step
    , form : BuildForm.Context
    , formModalVisibility : Modal.Visibility
    , selectedTab : Maybe Tab
    , frame : Frame
    }


type alias FromBuild =
    Build.Id


type alias ToBuild =
    Build.Id


type Stream
    = Stream BuildStream.Id


type Tab
    = BuildTab Int


type Frame
    = BuildFrame BuildType
    | BlankFrame


type BuildType
    = LoadedBuild Build.Id BuildOutput.Model
    | LoadingBuild (Maybe FromBuild) (Maybe ToBuild)


stringToTab : Maybe String -> List Build -> Maybe Tab
stringToTab maybeSelectedTab builds =
    case maybeSelectedTab of
        Just tabText ->
            tabText
                |> String.split "-"
                |> List.reverse
                |> List.head
                |> Maybe.andThen (String.toInt >> Result.toMaybe)
                |> Maybe.map BuildTab

        Nothing ->
            Nothing


init : Context -> Session msg -> Project.Id -> Commit.Hash -> ProjectTask.Task -> Maybe String -> List Build -> Task PageLoadError Model
init context session id hash task maybeSelectedTab builds =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        selectedTab =
            stringToTab maybeSelectedTab builds

        maybeMostRecentBuild =
            builds
                |> List.reverse
                |> List.head

        init =
            maybeBuildToModel context task maybeAuthToken selectedTab
    in
        case selectedTab of
            Just (BuildTab buildIndex) ->
                let
                    build =
                        builds
                            |> Array.fromList
                            |> Array.get (buildIndex - 1)
                in
                    init build

            Nothing ->
                init maybeMostRecentBuild


maybeBuildToModel :
    Context
    -> ProjectTask.Task
    -> Maybe AuthToken
    -> Maybe Tab
    -> Maybe Build
    -> Task PageLoadError Model
maybeBuildToModel context task maybeAuthToken selectedTab maybeBuild =
    let
        init =
            initialModel task selectedTab
    in
        case maybeBuild of
            Just b ->
                BuildOutput.init context task maybeAuthToken b
                    |> Task.map (LoadedBuild b.id >> BuildFrame >> init)
                    |> Task.mapError handleLoadError

            Nothing ->
                Task.succeed (init BlankFrame)


initialModel : ProjectTask.Task -> Maybe Tab -> Frame -> Model
initialModel task selectedTab frame =
    { task = task
    , toggledStep = Nothing
    , form = BuildForm.init task
    , formModalVisibility = Modal.hidden
    , selectedTab = selectedTab
    , frame = frame
    }


handleLoadError : a -> PageLoadError
handleLoadError _ =
    pageLoadError Page.Project "Project unavailable."



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions { formModalVisibility } =
    Modal.subscriptions formModalVisibility AnimateFormModal



-- CHANNELS --


streamChannelName : BuildStream -> String
streamChannelName stream =
    "stream:" ++ (BuildStream.idToString stream.id)


events : Model -> Dict String (List ( String, Encode.Value -> Msg ))
events model =
    case model.frame of
        BuildFrame (LoadedBuild _ buildOutputModel) ->
            BuildOutput.events buildOutputModel
                |> mapEvents BuildOutputMsg

        _ ->
            Dict.empty


leaveChannels : Model -> List String
leaveChannels model =
    case model.frame of
        BuildFrame (LoadedBuild _ buildOutputModel) ->
            BuildOutput.leaveChannels buildOutputModel

        _ ->
            []


mapEvents :
    (b -> c)
    -> Dict comparable (List ( a1, a -> b ))
    -> Dict comparable (List ( a1, a -> c ))
mapEvents fromMsg events =
    events
        |> Dict.map (\_ v -> List.map (Tuple.mapSecond (\msg -> msg >> fromMsg)) v)



-- VIEW --


view : Project -> Commit -> Model -> List Build -> Html Msg
view project commit model builds =
    let
        task =
            model.task

        stepList =
            viewStepList task.steps model.toggledStep

        navigation =
            if List.isEmpty builds then
                text ""
            else
                model.selectedTab
                    |> Maybe.map (viewTabs project commit task builds)
                    |> Maybe.withDefault (text "")
    in
        div [ class "row" ]
            [ div [ class "col-sm-12 col-md-12 col-lg-12 default-margin-bottom" ]
                [ viewToolbar
                , navigation
                , viewTabFrame model builds
                , viewFormModal model.task model.form model.formModalVisibility
                ]
            ]


viewFormModal : ProjectTask.Task -> BuildForm.Context -> Modal.Visibility -> Html Msg
viewFormModal task form visibility =
    let
        hasFields =
            not (List.isEmpty form.fields)

        basicModal =
            Modal.config CloseFormModal
                |> Modal.withAnimation AnimateFormModal
                |> Modal.large
                |> Modal.hideOnBackdropClick True
                |> Modal.h3 [] [ text (ProjectTask.nameToString task.name) ]
                |> Modal.footer [] [ BuildForm.viewSubmitButton buildFormConfig form ]

        modal =
            if hasFields then
                Modal.body [] (BuildForm.view buildFormConfig form) basicModal
            else
                basicModal
    in
        Modal.view visibility modal


viewTabs : Project -> Commit -> ProjectTask.Task -> List Build -> Tab -> Html Msg
viewTabs project commit task builds selectedTab =
    let
        compare a b =
            case ( a, b ) of
                ( BuildTab c, BuildTab d ) ->
                    c == d

        buildTab t =
            let
                build =
                    case t of
                        BuildTab b ->
                            Array.fromList builds
                                |> Array.get (b - 1)

                tabContent =
                    case t of
                        BuildTab b ->
                            text ("Build #" ++ (toString b) ++ " ")

                tabQueryParam =
                    case t of
                        BuildTab b ->
                            "build-" ++ (toString b)

                route =
                    Just tabQueryParam
                        |> CommitRoute.Task task.name
                        |> ProjectRoute.Commit commit.hash
                        |> Route.Project project.slug

                tabIcon =
                    build
                        |> Maybe.map viewBuildStatusIcon
                        |> Maybe.withDefault (text "")

                textClass =
                    build
                        |> Maybe.map viewBuildTextClass
                        |> Maybe.withDefault ("")

                tabClassList =
                    [ ( "nav-link", True )
                    , ( textClass, True )
                    , ( "active", compare t selectedTab )
                    ]
            in
                li [ class "nav-item" ]
                    [ a
                        [ classList tabClassList
                        , Route.href route
                        , onClickPage (SelectTab selectedTab) route
                        ]
                        [ tabContent
                        , text " "
                        , tabIcon
                        ]
                    ]

        buildTabs =
            builds
                |> List.indexedMap (\i -> (\_ -> buildTab (BuildTab (i + 1))))
    in
        ul [ class "nav nav-tabs nav-fill" ] buildTabs


viewToolbar : Html Msg
viewToolbar =
    div [ class "btn-toolbar d-flex flex-row-reverse" ]
        [ button
            [ class "btn btn-primary btn-lg"
            , style [ "border-radius" => "25px" ]
            , onClick ShowOrSubmitTaskForm
            ]
            [ i [ class "fa fa-plus" ] [] ]
        ]


viewTabFrame : Model -> List Build -> Html Msg
viewTabFrame model builds =
    let
        findBuild id =
            builds
                |> List.filter (\a -> a.id == id)
                |> List.head
    in
        case model.frame of
            BlankFrame ->
                text ""

            BuildFrame (LoadedBuild buildId buildOutputModel) ->
                case findBuild buildId of
                    Just build ->
                        BuildOutput.view build buildOutputModel
                            |> Html.map BuildOutputMsg

                    Nothing ->
                        text ""

            BuildFrame (LoadingBuild _ _) ->
                text ""


breadcrumb : Project -> Commit -> ProjectTask.Task -> List ( Route.Route, String )
breadcrumb project commit task =
    [ ( CommitRoute.Task task.name Nothing |> ProjectRoute.Commit commit.hash |> Route.Project project.slug
      , ProjectTask.nameToString task.name
      )
    ]



-- UPDATE --


type Msg
    = ToggleStep (Maybe Step)
    | OnInput BuildForm.InputFormField String
    | OnChange BuildForm.ChoiceFormField (Maybe Int)
    | SubmitForm
    | BuildCreated (Result Request.Errors.HttpError Build)
    | SelectTab Tab String
    | BuildLoaded (Result Request.Errors.HttpError (Maybe BuildType))
    | BuildOutputMsg BuildOutput.Msg
    | CloseFormModal
    | AnimateFormModal Modal.Visibility
    | ShowOrSubmitTaskForm


type ExternalMsg
    = NoOp
    | AddBuild Build
    | UpdateBuild Build


buildFormConfig : BuildForm.Config Msg
buildFormConfig =
    { submitMsg = SubmitForm
    , onChangeMsg = OnChange
    , onInputMsg = OnInput
    }


update : Context -> Project -> Commit -> List Build -> Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update context project commit builds session msg model =
    let
        projectSlug =
            project.slug

        commitHash =
            commit.hash

        taskName =
            model.task.name

        maybeAuthToken =
            Maybe.map .token session.user
    in
        case msg of
            ShowOrSubmitTaskForm ->
                let
                    form =
                        BuildForm.init model.task

                    hasFields =
                        not (List.isEmpty form.fields)
                in
                    if hasFields then
                        { model
                            | formModalVisibility = Modal.shown
                            , form = form
                        }
                            => Cmd.none
                            => NoOp
                    else
                        update context project commit builds session SubmitForm model

            CloseFormModal ->
                { model
                    | formModalVisibility = Modal.hidden
                    , form = BuildForm.init model.task
                }
                    => Cmd.none
                    => NoOp

            AnimateFormModal visibility ->
                { model | formModalVisibility = visibility }
                    => Cmd.none
                    => NoOp

            ToggleStep maybeStep ->
                { model | toggledStep = maybeStep }
                    => Cmd.none
                    => NoOp

            OnInput field value ->
                { model | form = BuildForm.updateInput field value model.form }
                    => Cmd.none
                    => NoOp

            OnChange field maybeIndex ->
                { model | form = BuildForm.updateSelect field maybeIndex model.form }
                    => Cmd.none
                    => NoOp

            SubmitForm ->
                let
                    cmdFromAuth authToken =
                        authToken
                            |> Request.Commit.createBuild context projectSlug commitHash taskName (BuildForm.submitParams model.form)
                            |> Task.attempt BuildCreated

                    cmd =
                        session
                            |> Session.attempt "create build" cmdFromAuth
                            |> Tuple.second
                in
                    model
                        => cmd
                        => NoOp

            BuildLoaded (Ok (Just loadedBuild)) ->
                model
                    => Cmd.none
                    => NoOp

            BuildLoaded _ ->
                model
                    => Cmd.none
                    => NoOp

            BuildCreated (Ok build) ->
                let
                    tabNum =
                        List.length builds

                    tab =
                        tabNum
                            |> toString
                            |> (\i -> "build-" ++ i)
                            |> Just

                    route =
                        CommitRoute.Task model.task.name tab
                            |> ProjectRoute.Commit commit.hash
                            |> Route.Project project.slug
                in
                    model
                        => Navigation.newUrl (Route.routeToString route)
                        => AddBuild build

            BuildCreated (Err _) ->
                model
                    => Cmd.none
                    => NoOp

            SelectTab tab url ->
                let
                    frame =
                        case tab of
                            BuildTab toBuildIndex ->
                                let
                                    toBuild =
                                        builds
                                            |> Array.fromList
                                            |> Array.get (toBuildIndex - 1)
                                            |> Maybe.map .id

                                    fromBuild =
                                        case model.frame of
                                            BuildFrame (LoadedBuild b _) ->
                                                Just b

                                            _ ->
                                                Nothing
                                in
                                    BuildFrame (LoadingBuild fromBuild toBuild)
                in
                    { model
                        | selectedTab = Just tab
                        , frame = frame
                    }
                        => Navigation.newUrl url
                        => NoOp

            BuildOutputMsg subMsg ->
                case model.frame of
                    BuildFrame (LoadedBuild id outputModel) ->
                        let
                            ( newOutputModel, newOutputCmd ) =
                                BuildOutput.update subMsg outputModel
                        in
                            { model | frame = BuildFrame (LoadedBuild id newOutputModel) }
                                => Cmd.map BuildOutputMsg newOutputCmd
                                => NoOp

                    _ ->
                        model
                            => Cmd.none
                            => NoOp
