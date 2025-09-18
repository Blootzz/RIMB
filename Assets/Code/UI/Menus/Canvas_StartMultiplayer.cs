using UnityEngine;
using Unity.Netcode;
using UnityEngine.SceneManagement;

public class Canvas_StartMultiplayer : MonoBehaviour
{
    public void StartHost()
    {
        NetworkManager.Singleton.StartHost();
        DeactivateCanvas();
    }
    public void StartClient()
    {
        NetworkManager.Singleton.StartClient();
        DeactivateCanvas();
    }

    void DeactivateCanvas()
    {
        print("Temporary solution. Consider unloading this scene or make this a prefab later");
    }
}
