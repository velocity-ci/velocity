port module Subscriptions exposing (Subscriptions)

import Graphql.Operation exposing (RootSubscription)
import Graphql.Document
import KnownHost exposing (KnownHost)
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)
import Api.Compiled.Subscription
import Dict exposing (Dict)


type Subscriptions
    = Subscriptions Internals


type alias Internals =
    { subscriptions : Dict Int Subscription
    }


type alias Subscription =
    { status : Status
    , subscriptionType : SubscriptionType
    }


port subscribeTo : List ( String, String ) -> Cmd msg


type Status
    = NotConnected
    | Connected
    | Reconnecting


type SubscriptionType
    = KnownHostSubscription (SelectionSet KnownHost RootSubscription)


init : Subscriptions
init =
    Subscriptions
        { subscriptions = Dict.empty
        }


subscribe : Subscriptions -> SubscriptionType -> ( Subscriptions, Cmd msg )
subscribe (Subscriptions internals) subscriptionType =
    let
        subscription =
            { status = NotConnected
            , subscriptionType = subscriptionType
            }

        id =
            (Dict.size internals.subscriptions) + 1
    in
        ( Subscriptions { internals | subscriptions = Dict.insert id subscription internals.subscriptions }
        , Cmd.none
        )


subscribe2 : Subscriptions -> SelectionSet decodesTo RootSubscription -> ( Subscriptions, Cmd msg )
subscribe2 (Subscriptions internals) subscriptionType =
    let
        subscription =
            { status = NotConnected
            , subscriptionType = subscriptionType
            }

        id =
            (Dict.size internals.subscriptions) + 1
    in
        ( Subscriptions { internals | subscriptions = Dict.insert id subscription internals.subscriptions }
        , Cmd.none
        )


knownHostAddedSubscription : SubscriptionType
knownHostAddedSubscription =
    KnownHostSubscription <|
        Api.Compiled.Subscription.knownHostAdded KnownHost.selectionSet


knownHostVerifiedSubscription : SubscriptionType
knownHostVerifiedSubscription =
    KnownHostSubscription <|
        Api.Compiled.Subscription.knownHostVerified KnownHost.selectionSet
