port module Subscriptions exposing (Subscriptions)

import Api.Compiled.Subscription
import Dict exposing (Dict)
import Graphql.Document as Document
import Graphql.Operation exposing (RootSubscription)
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)
import KnownHost exposing (KnownHost)



-- OUTWARDS PORT


port subscribeTo : ( Int, String ) -> Cmd msg



-- TYPES


type Subscriptions msg
    = Subscriptions (Dict Int (SubscriptionType msg))


type SubscriptionType msg
    = KnownHostSubscription Status (SelectionSet KnownHost RootSubscription) (KnownHost -> msg)


type alias Subscription msg =
    { status : Status
    , subscriptionType : SubscriptionType msg
    }


type Status
    = NotConnected
    | Connected
    | Reconnecting


init : Subscriptions msg
init =
    Subscriptions Dict.empty


subscribeToKnownHostAdded : (KnownHost -> msg) -> Subscriptions msg -> ( Subscriptions msg, Cmd msg )
subscribeToKnownHostAdded toMsg subscriptions =
    Api.Compiled.Subscription.knownHostAdded KnownHost.selectionSet
        |> knownHostSubscription toMsg subscriptions


knownHostSubscription :
    (KnownHost -> msg)
    -> Subscriptions msg
    -> SelectionSet KnownHost RootSubscription
    -> ( Subscriptions msg, Cmd msg )
knownHostSubscription toMsg (Subscriptions subscriptions) selectionSet =
    let
        sub =
            KnownHostSubscription NotConnected selectionSet toMsg

        nextId =
            Dict.size subscriptions
    in
    ( Subscriptions (Dict.insert nextId sub subscriptions)
    , subscribeTo ( nextId, Document.serializeSubscription selectionSet )
    )
