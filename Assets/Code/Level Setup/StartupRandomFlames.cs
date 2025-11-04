using Unity.Netcode;
using Unity.VisualScripting;
using UnityEngine;

public class StartupRandomFlames : NetworkBehaviour
{
    [InspectorButton(nameof(OnButtonClicked))]
    public bool spawnFlames;
    private void OnButtonClicked() { SpawnFlames(); }

    [SerializeField] GameObject flamePrefab;

    [Header("Spawn Settings")]
    [SerializeField] float numFlames;
    [SerializeField][Range(0, 100)] float maxDistanceX;
    [SerializeField] float posY;
    [SerializeField][Range(0, 100)] float maxDistanceZ;

    public void SpawnFlames()
    {
        if (!IsHost)
            return;

        Vector3 spawnPos;
        for (int i = 0; i < numFlames; i++)
        {
            spawnPos.x = Random.value * maxDistanceX / 2;
            spawnPos.y = posY;
            spawnPos.z = Random.value * maxDistanceZ / 2;

            var localInstance = Instantiate(flamePrefab, spawnPos, Quaternion.identity);
            localInstance.GetComponent<NetworkObject>().Spawn(/*true*/);
        }
    }
}
