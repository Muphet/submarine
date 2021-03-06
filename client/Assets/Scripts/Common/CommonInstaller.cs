﻿using Zenject;
using Type = TyphenApi.Type.Submarine;

namespace Submarine
{
    public class CommonInstaller : MonoInstaller
    {
        public override void InstallBindings()
        {
            Container.Bind<SignalManager>().AsSingle();

            Container.Bind<Type.Configuration.Client>().FromInstance(Type.Configuration.Client.Load()).AsSingle();
            Container.Bind<TyphenApi.WebApi.Submarine>().AsSingle();

            Container.Bind<UserModel>().AsSingle();
            Container.Bind<LobbyModel>().AsSingle();
            Container.Bind<PermanentDataStoreService>().AsSingle();

            Container.DeclareSignal<SceneChangeCommand>().RequireHandler();
            Container.BindSignal<SceneNames, SceneChangeCommand>().To<SceneChangeCommand.Handler>(x => x.Execute).AsSingle();

            Container.BindInterfacesAndSelfTo<CommonMediator>().AsSingle();
        }
    }
}
