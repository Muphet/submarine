// This file was generated by typhen-api

using System;
using System.Collections;
using System.Collections.Generic;

namespace TyphenApi
{
    public class SynchronizedFunctionCaller
    {
        List<Action> reservedFunctions = new List<Action>();

        public void ReserveCall(Action function)
        {
            lock (((ICollection)reservedFunctions).SyncRoot)
            {
                reservedFunctions.Add(function);
            }
        }

        public void Call()
        {
            Action[] functions;

            lock (((ICollection)reservedFunctions).SyncRoot)
            {
                functions = reservedFunctions.ToArray();
                reservedFunctions.Clear();
            }

            foreach (var function in functions)
            {
                function();
            }
        }
    }
}
